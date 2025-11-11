using AbyssCLI.Tool;

namespace AbyssCLI.Client;

public static partial class Client
{
    public static void SwapMainWorld(AbyssURL url) //can also be called from javascript API.
    {
        lock (_world_move_lock)
        {
            AbyssLib.World net_world;
            AbyssURL world_url;
            if (url.Scheme == "abyss")
            {
                net_world = Host.JoinWorld(url.Raw);
                if (!net_world.IsValid())
                {
                    CerrWriteLine("failed to join world: " + url.Raw);
                    return;
                }
                if (!AbyssURLParser.TryParse(net_world.url, out world_url) || world_url.Scheme == "abyss")
                {
                    CerrWriteLine("invalid world url: " + world_url.Raw);
                    _ = net_world.Leave();
                    return;
                }
            }
            else
            {
                net_world = Host.OpenWorld(url.Raw);
                world_url = url;
            }

            if (!net_world.IsValid())
            {
                CerrWriteLine("MoveWorld: failed to open world");
                return;
            }

            _ = _resolver.DeleteMapping("");
            _current_world?.Leave();
            try
            {
                _current_world = new World(Host, net_world, world_url);
            }
            catch (Exception ex)
            {
                CerrWriteLine("world creation failed: " + ex.Message);
                _current_world = null;
            }

            if (!_resolver.TrySetMapping("", net_world.world_id).Empty)
            {
                throw new Exception("failed to set world path mapping");
            }
        }
    }
}
