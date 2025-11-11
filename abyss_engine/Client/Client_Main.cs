using AbyssCLI.ABI;
using System.Collections.Concurrent;
using System.Threading.Channels;

namespace AbyssCLI.Client;

public partial class Client
{
    private static readonly Channel<UIAction> _HL_Action_Channel = Channel.CreateUnbounded<UIAction>(new UnboundedChannelOptions
    {
        SingleReader = true,
        SingleWriter = false
    });

    private static UIAction ReadProtoMessage()
    {
        int length = _cin.ReadInt32();
        byte[] data = _cin.ReadBytes(length);
        if (data.Length != length)
        {
            throw new Exception("stream closed");
        }
        return UIAction.Parser.ParseFrom(data);
    }

    public static async Task IssueMoveWorldInternalRequest(string world_url)
    {
        var message = new UIAction()
        {
            MoveWorld = new()
            {
                WorldUrl = world_url
            }
        };
        await _HL_Action_Channel.Writer.WriteAsync(message);
    }

    private static async Task<bool> UIActionHandle()
    {
        UIAction message = await _HL_Action_Channel.Reader.ReadAsync();
        switch (message.InnerCase)
        {
        case UIAction.InnerOneofCase.Kill:
            return false;
        case UIAction.InnerOneofCase.MoveWorld:
            OnMoveWorld(message.MoveWorld);
            return true;
        case UIAction.InnerOneofCase.ShareContent:
            OnShareContent(message.ShareContent);
            return true;
        case UIAction.InnerOneofCase.UnshareContent:
            OnUnshareContent(message.UnshareContent);
            return true;
        case UIAction.InnerOneofCase.ConnectPeer:
            OnConnectPeer(message.ConnectPeer);
            return true;
        case UIAction.InnerOneofCase.ConsoleInput:
            OnConsoleInput(message.ConsoleInput);
            return true;
        default:
            throw new Exception("fatal: received invalid UI Action");
        }
    }

    public static async Task Run()
    {
        _=Task.Run(async () =>
        {
            while (true)
            {
                var message = ReadProtoMessage();
                await _HL_Action_Channel.Writer.WriteAsync(message);
                if (message.InnerCase == UIAction.InnerOneofCase.Kill)
                {
                    break;
                }
            }
        });

        while (await UIActionHandle())
        {
        }
    }
}
