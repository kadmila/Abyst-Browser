namespace AbyssCLI.Test;

public static class ExternalDllTest
{
    public static void TestDllLoad()
    {
        Console.WriteLine(AbyssLib.GetVersion());
        Console.WriteLine(AbyssLib.GetError().ToString());
    }

    public static void TestHostCreate()
    {
        byte[] priv_key = File.ReadAllBytes("test_key1.pem");
        AbyssLib.SimplePathResolver path_res = AbyssLib.NewSimplePathResolver();
        AbyssLib.Host host = AbyssLib.OpenAbyssHost(priv_key, path_res, AbyssLib.NewSimpleAbystServer("D:\\WORKS\\github\\abyss_engine\\testground\\abyst_server"));
        Console.WriteLine(host.IsValid() ? "success" : "fail");
    }
    public static void TestHostJoin()
    {
        byte[] priv_key1 = File.ReadAllBytes("test_key1.pem");
        byte[] priv_key2 = File.ReadAllBytes("test_key2.pem");

        AbyssLib.SimplePathResolver path_res1 = AbyssLib.NewSimplePathResolver();
        AbyssLib.SimplePathResolver path_res2 = AbyssLib.NewSimplePathResolver();

        AbyssLib.Host host1 = AbyssLib.OpenAbyssHost(priv_key1, path_res1, AbyssLib.NewSimpleAbystServer("D:\\WORKS\\github\\abyss_engine\\testground\\abyst_server"));
        AbyssLib.Host host2 = AbyssLib.OpenAbyssHost(priv_key2, path_res2, AbyssLib.NewSimpleAbystServer("D:\\WORKS\\github\\abyss_engine\\testground\\abyst_server"));

        _ = host1.AppendKnownPeer(host2.root_certificate, host2.handshake_key_certificate);
        _ = host2.AppendKnownPeer(host1.root_certificate, host1.handshake_key_certificate);

        AbyssLib.World world1 = host1.OpenWorld("plain.world");
        if (!world1.IsValid())
        {
            Console.WriteLine("failed to open world");
            return;
        }
        _ = path_res1.TrySetMapping("/cat", world1.world_id);

        Thread.Sleep(1000);

        var host1_th = new Thread(() =>
        {
            AbyssLib.ErrorCode err = host1.OpenOutboundConnection(host2.local_aurl.Raw);

            dynamic evnt_raw = world1.WaitForEvent();
            {
                var evnt = evnt_raw as AbyssLib.WorldMemberRequest;
                Console.WriteLine("host2: " + evnt.peer_hash);
                _ = evnt.Accept();
            }
            evnt_raw = world1.WaitForEvent();
            {
                if (evnt_raw is AbyssLib.WorldMember evnt)
                {
                    Console.WriteLine("Success(1)!");
                }
            }
        });
        host1_th.Start();

        AbyssLib.World world = host2.JoinWorld(host1.local_aurl + "cat");
        if (world.IsValid())
        {
            dynamic evnt_raw = world.WaitForEvent();
            {
                var evnt = evnt_raw as AbyssLib.WorldMemberRequest;
                Console.WriteLine("host1: " + evnt.peer_hash);
                _ = evnt.Accept();
            }
            evnt_raw = world.WaitForEvent();
            {
                if (evnt_raw is AbyssLib.WorldMember evnt)
                {
                    Console.WriteLine("Success(2)!");
                }
            }
        }
        else
        {
            Console.WriteLine("failed to join world");
        }

        host1_th.Join();
    }
    public static void TestObjectSharing()
    {
        byte[] priv_key1 = File.ReadAllBytes("test_key1.pem");
        byte[] priv_key2 = File.ReadAllBytes("test_key2.pem");

        AbyssLib.SimplePathResolver path_res1 = AbyssLib.NewSimplePathResolver();
        AbyssLib.SimplePathResolver path_res2 = AbyssLib.NewSimplePathResolver();

        AbyssLib.Host host1 = AbyssLib.OpenAbyssHost(priv_key1, path_res1, AbyssLib.NewSimpleAbystServer("D:\\WORKS\\github\\abyss_engine\\testground\\abyst_server"));
        AbyssLib.Host host2 = AbyssLib.OpenAbyssHost(priv_key2, path_res2, AbyssLib.NewSimpleAbystServer("D:\\WORKS\\github\\abyss_engine\\testground\\abyst_server"));

        _ = host1.AppendKnownPeer(host2.root_certificate, host2.handshake_key_certificate);
        _ = host2.AppendKnownPeer(host1.root_certificate, host1.handshake_key_certificate);

        AbyssLib.World world1 = host1.OpenWorld("plain.world");
        _ = path_res1.TrySetMapping("/cat", world1.world_id);

        Thread.Sleep(1000);

        new Thread(() =>
        {
            AbyssLib.ErrorCode err = host1.OpenOutboundConnection(host2.local_aurl.Raw);
            _ = (world1.WaitForEvent() as AbyssLib.WorldMemberRequest).Accept();
            var mem = world1.WaitForEvent() as AbyssLib.WorldMember;
            Console.WriteLine("joined(1)!");

            var obj_id = Guid.NewGuid();
            _ = mem.AppendObjects([Tuple.Create(obj_id, "thats.com/carrot.aml", new float[] { 0.0f, 0.0f, 0.0f, 1.0f, 0.0f, 0.0f, 0.0f })]);
            Console.WriteLine("Object appended!");

            Thread.Sleep(500);

            _ = mem.DeleteObjects([obj_id]);
            Console.WriteLine("Object deleted!");
        }).Start();

        AbyssLib.World world = host2.JoinWorld(host1.local_aurl + "cat");
        _ = (world.WaitForEvent() as AbyssLib.WorldMemberRequest).Accept();
        var mem = world.WaitForEvent() as AbyssLib.WorldMember;
        Console.WriteLine("joined(2)!");

        var oap_ev = world.WaitForEvent() as AbyssLib.MemberObjectAppend;
        if (oap_ev.peer_hash != mem.hash)
        {
            Console.WriteLine("member mismatch");
            return;
        }
        Console.WriteLine("received object: " + oap_ev.objects[0].Item2);
        var ode_ev = world.WaitForEvent() as AbyssLib.MemberObjectDelete;
        Console.WriteLine("deleted object: " + ode_ev.object_ids[0]);

        if (oap_ev.objects[0].Item1 != ode_ev.object_ids[0])
        {
            Console.WriteLine("object not same!");
        }
    }
}
