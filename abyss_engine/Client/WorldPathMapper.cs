namespace AbyssCLI.Client;

internal class WorldPathMapper
{
    public bool TryAddMapping(string localpath, string world_uuid, World world)
    {
        lock (_lock)
        {
            if (!_path_map.TryAdd(localpath, world_uuid))
                return false;

            if (!_world_map.TryAdd(world_uuid, world))
            {
                _ = _path_map.Remove(localpath);
                return false;
            }

            return true;
        }
    }
    public bool TryPopMapping(string localpath, out World world)
    {
        world = null;
        lock (_lock)
        {
            if (!_path_map.TryGetValue(localpath, out string world_uuid))
                return false;

            _ = _path_map.Remove(localpath);
            world = _world_map[world_uuid];
            _ = _world_map.Remove(world_uuid); //this must success.
            return true;
        }
    }
    public bool TryGetUUID(string localpath, out string world_uuid)
    {
        lock (_lock)
        {
            return _path_map.TryGetValue(localpath, out world_uuid);
        }
    }
    public bool TryGetWorld(string world_uuid, out World world)
    {
        lock (_lock)
        {
            return _world_map.TryGetValue(world_uuid, out world);
        }
    }

    private readonly object _lock = new();
    private readonly Dictionary<string, string> _path_map = []; //localpath - uuid
    private readonly Dictionary<string, World> _world_map = []; //uuid - world
}
