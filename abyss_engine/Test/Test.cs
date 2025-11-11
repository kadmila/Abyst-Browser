//using System.Text.Json;

//namespace AbyssCLI.Test
//{
//    internal class ClientTest
//    {
//        public ClientTest()
//        {
//            Console.WriteLine(AbyssLib.GetVersion());
//            var hostA = new AbyssLib.AbyssHost("mallang_host_A", "D:\\WORKS\\github\\abyss\\temp");
//            Console.WriteLine(hostA.LocalAddr());

//            var hostB = new AbyssLib.AbyssHost("mallang_host_B", "D:\\WORKS\\github\\abyss\\temp");
//            Console.WriteLine(hostB.LocalAddr());

//            var hostC = new AbyssLib.AbyssHost("mallang_host_C", "D:\\WORKS\\github\\abyss\\temp");
//            Console.WriteLine(hostC.LocalAddr());

//            var hostD = new AbyssLib.AbyssHost("mallang_host_D", "D:\\WORKS\\github\\abyss\\temp");
//            Console.WriteLine(hostD.LocalAddr());

//            //A
//            new Thread(() =>
//            {
//                while (true)
//                {
//                    var ev_a = hostA.AndWaitEvent();
//                    Console.WriteLine("A: " + JsonSerializer.Serialize(ev_a));
//                }
//            }).Start();
//            new Thread(() =>
//            {
//                while (true)
//                {
//                    Console.WriteLine("Err[A] " + hostA.WaitError().ToString());
//                }
//            }).Start();
//            new Thread(() =>
//            {
//                while (true)
//                {
//                    Console.WriteLine("SOM[A] " + hostA.SomWaitEvent().ToString());
//                }
//            }).Start();

//            //B
//            new Thread(() =>
//            {
//                while (true)
//                {
//                    var ev_b = hostB.AndWaitEvent();
//                    Console.WriteLine("B: " + JsonSerializer.Serialize(ev_b));
//                }
//            }).Start();
//            new Thread(() =>
//            {
//                while (true)
//                {
//                    Console.WriteLine("Err[B] " + hostB.WaitError().ToString());
//                }
//            }).Start();
//            new Thread(() =>
//            {
//                while (true)
//                {
//                    Console.WriteLine("SOM[A] " + hostB.SomWaitEvent().ToString());
//                }
//            }).Start();

//            //C
//            new Thread(() =>
//            {
//                while (true)
//                {
//                    var ev_c = hostC.AndWaitEvent();
//                    Console.WriteLine("C: " + JsonSerializer.Serialize(ev_c));
//                }
//            }).Start();
//            new Thread(() =>
//            {
//                while (true)
//                {
//                    Console.WriteLine("Err[C] " + hostC.WaitError().ToString());
//                }
//            }).Start();
//            new Thread(() =>
//            {
//                while (true)
//                {
//                    Console.WriteLine("SOM[A] " + hostC.SomWaitEvent().ToString());
//                }
//            }).Start();

//            //D
//            new Thread(() =>
//            {
//                while (true)
//                {
//                    var ev_d = hostD.AndWaitEvent();
//                    Console.WriteLine("D: " + JsonSerializer.Serialize(ev_d));
//                }
//            }).Start();
//            new Thread(() =>
//            {
//                while (true)
//                {
//                    Console.WriteLine("Err[D] " + hostD.WaitError().ToString());
//                }
//            }).Start();
//            new Thread(() =>
//            {
//                while (true)
//                {
//                    Console.WriteLine("SOM[A] " + hostD.SomWaitEvent().ToString());
//                }
//            }).Start();

//            Console.WriteLine("----");

//            hostA.AndOpenWorld("/", "https://www.abysseum.com/");
//            Thread.Sleep(1000);

//            hostA.RequestConnect(hostB.LocalAddr());
//            hostB.AndJoin("/", hostA.LocalAddr());
//            Thread.Sleep(1000);

//            hostB.RequestConnect(hostC.LocalAddr());
//            hostC.AndJoin("/", hostB.LocalAddr());
//            Thread.Sleep(1000);

//            hostC.RequestConnect(hostD.LocalAddr());
//            hostD.AndJoin("/", hostC.LocalAddr());

//            Thread.Sleep(2000);

//            //var response = hostB.HttpGet("abyst://mallang_host_D/static/key.pem");
//            //Console.WriteLine("response(" + response.GetStatus() + ")" + Encoding.UTF8.GetString(response.GetBody()));

//            //TODO: fix
//            Console.WriteLine("host A closing world /");
//            hostA.AndCloseWorld("/");
//            Thread.Sleep(2000);

//            Console.WriteLine("host A joining C world /");
//            hostA.AndJoin("/", hostC.LocalAddr());

//            Thread.Sleep(5000);

//            Console.WriteLine("host B sharing object /");
//            //hostB.SomInitiateService("hostA", )
//        }
//    }
//}
