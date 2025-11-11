#nullable enable
#pragma warning disable IDE1006 //naming convension
namespace AbyssCLI.AML.JavaScriptAPI;
public class Host
{
    public string aurl => Client.Client.Host.local_aurl.Raw;
    public string id => Client.Client.Host.local_aurl.Id;
    public string idCert => System.Text.Encoding.UTF8.GetString(Client.Client.Host.root_certificate);
    public string hsKeyCert => System.Text.Encoding.UTF8.GetString(Client.Client.Host.handshake_key_certificate);
    public void register(string id_cert, string hs_key_cert)
    {
        var result = Client.Client.Host.AppendKnownPeer(System.Text.Encoding.UTF8.GetBytes(id_cert), System.Text.Encoding.UTF8.GetBytes(hs_key_cert));
        if (result != null)
        {
            Client.Client.RenderWriter.ConsolePrint("register failed: " + result.Message);
        }
    }
    public void connect(string aurl)
    {
        var result = Client.Client.Host.OpenOutboundConnection(aurl);
        if (result != AbyssLib.ErrorCode.SUCCESS)
        {
            Client.Client.RenderWriter.ConsolePrint("connect failed: " + result.ToString());
        }
    }
    public void move_world(string aurl)
    {
        _ = Client.Client.IssueMoveWorldInternalRequest(aurl);
    }
}
