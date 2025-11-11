using AbyssCLI.Tool;
using System.Numerics;

namespace AbyssCLI.Client;

public class World
{
    private readonly AbyssLib.Host _host;
    private readonly AbyssLib.World _world;
    internal readonly HL.ContentB _environment;
    private readonly Dictionary<string, HL.Member> _members = []; //peer hash - [uuid - item]
    private readonly Dictionary<Guid, HL.Item> _local_items = []; //UUID - item
    private readonly object _lock = new();
    private readonly Thread _world_th;

    public World(AbyssLib.Host host, AbyssLib.World world, AbyssURL URL)
    {
        _host = host;
        _world = world;
        _environment = new(URL, new()
        {
            title = URL.ToString()
        });
        _world_th = new Thread(() =>
        {
            while (true)
            {
                dynamic evnt_raw = world.WaitForEvent();
                switch (evnt_raw)
                {
                case AbyssLib.WorldMemberRequest evnt:
                    OnMemberRequest(evnt);
                    break;
                case AbyssLib.WorldMember evnt:
                    OnMemberReady(evnt);
                    break;
                case AbyssLib.MemberObjectAppend evnt:
                    OnMemberObjectAppend(evnt);
                    break;
                case AbyssLib.MemberObjectDelete evnt:
                    OnMemberObjectDelete(evnt);
                    break;
                case AbyssLib.WorldMemberLeave evnt:
                    OnMemberLeave(evnt.peer_hash);
                    break;
                case int: //world termination
                    return;
                }
            }
        });
        _world_th.Start();
    }

    public void ShareItem(Guid uuid, AbyssURL url, float[] transform)
    {
        var item = new HL.Item(_host.local_aurl.Id, uuid, url,
            new(transform[0], transform[1], transform[2]),
            new(transform[4], transform[5], transform[6], transform[3]));

        lock (_lock)
        {
            _local_items[uuid] = item;
            foreach (KeyValuePair<string, HL.Member> entry in _members)
            {
                _ = entry.Value.network_handle.AppendObjects([Tuple.Create(uuid, url.Raw, transform)]);
            }
        }
    }

    public void UnshareItem(Guid guid)
    {
        lock (_lock)
        {
            HL.Item item = _local_items[guid];
            item.Stop();
            _ = _local_items.Remove(guid);
            foreach (KeyValuePair<string, HL.Member> member in _members)
            {
                _ = member.Value.network_handle.DeleteObjects([guid]);
            }
        }
    }

    public void Leave()
    {
        _environment.Dispose();
        if (_world.Leave() != 0)
        {
            Client.CerrWriteLine("failed to leave world");
        }
        _world_th.Join();

        foreach (KeyValuePair<string, HL.Member> member in _members)
        {
            foreach (HL.Item item in member.Value.remote_items.Values)
            {
                item.Stop();
            }
        }
        foreach (HL.Item item in _local_items.Values)
        {
            item.Stop();
        }
        _members.Clear(); //do we need this?
        _local_items.Clear(); //do we need this?
    }

    //internals
    private static void OnMemberRequest(AbyssLib.WorldMemberRequest evnt)
    {
        Client.CerrWriteLine("OnMemberRequest");
        _ = evnt.Accept();
    }
    private void OnMemberReady(AbyssLib.WorldMember member)
    {
        Client.CerrWriteLine("OnMemberReady");
        lock (_lock)
        {
            if (!_members.TryAdd(member.hash, new(member)))
            {
                Client.CerrWriteLine("failed to append peer; old peer session pends");
                return;
            }
            Client.RenderWriter.MemberInfo(member.hash);

            static float[] PosRotSerialize(Vector3 pos, Quaternion rot) =>
                [pos.X, pos.Y, pos.Z, rot.W, rot.X, rot.Y, rot.Z];

            Tuple<Guid, string, float[]>[] list_of_local_items =
                [.. _local_items.Select(kvp =>
                    Tuple.Create(
                        kvp.Key,
                        kvp.Value._url.ToString(),
                        PosRotSerialize(
                            kvp.Value._content.Document.Metadata.pos.Native,
                            kvp.Value._content.Document.Metadata.rot.Native
                        )
                    )
                )];
            if (list_of_local_items.Length != 0)
            {
                _ = member.AppendObjects(list_of_local_items);
            }
        }
    }

    private void OnMemberObjectAppend(AbyssLib.MemberObjectAppend evnt)
    {
        Client.CerrWriteLine("OnMemberObjectAppend");
        var parsed_objects = evnt.objects
            .Select(gst =>
            {
                if (!AbyssURLParser.TryParse(gst.Item2, out AbyssURL abyss_url))
                {
                    Client.CerrWriteLine("failed to parse object url: " + gst.Item2);
                }
                return Tuple.Create(gst.Item1, abyss_url, gst.Item3);
            })
            .Where(gst => gst.Item2 != null)
            .ToList();

        lock (_lock)
        {
            if (!_members.TryGetValue(evnt.peer_hash, out HL.Member member))
            {
                Client.CerrWriteLine("failed to find member");
                return;
            }

            foreach (Tuple<Guid, AbyssURL, float[]> obj in parsed_objects)
            {
                Client.CerrWriteLine("member object: " + obj.Item2.ToString());
                var item = new HL.Item(evnt.peer_hash, obj.Item1, obj.Item2,
                    new(obj.Item3[0], obj.Item3[1], obj.Item3[2]),
                    new(obj.Item3[4], obj.Item3[5], obj.Item3[6], obj.Item3[3]));
                if (!member.remote_items.TryAdd(obj.Item1, item))
                {
                    Client.CerrWriteLine("uid collision of objects appended from peer");
                    continue;
                }
            }
        }
    }
    private void OnMemberObjectDelete(AbyssLib.MemberObjectDelete evnt)
    {
        Client.CerrWriteLine("OnMemberObjectDelete");
        lock (_lock)
        {
            if (!_members.TryGetValue(evnt.peer_hash, out HL.Member member))
            {
                Client.CerrWriteLine("failed to find member");
                return;
            }

            foreach (Guid id in evnt.object_ids)
            {
                if (!member.remote_items.Remove(id, out HL.Item item))
                {
                    Client.CerrWriteLine("peer tried to delete unshared objects");
                    continue;
                }
                item.Stop();
            }
        }
    }
    private void OnMemberLeave(string peer_hash)
    {
        Client.CerrWriteLine("OnMemberLeave");
        lock (_lock)
        {
            if (!_members.Remove(peer_hash, out HL.Member value))
            {
                Client.CerrWriteLine("non-existing peer leaved");
                return;
            }
            Client.RenderWriter.MemberLeave(peer_hash);

            foreach (HL.Item item in value.remote_items.Values)
            {
                item.Stop();
            }
        }
    }
}
