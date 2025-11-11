#nullable enable
#pragma warning disable IDE1006 //naming convension
namespace AbyssCLI.AML.JavaScriptAPI;

public class Document
{
    private readonly JavaScriptDispatcher _js_dispatcher;
    private readonly AML.Document _origin;
    internal Document(JavaScriptDispatcher js_dispatcher, AML.Document origin)
    {
        _js_dispatcher = js_dispatcher;
        _origin = origin;
    }
    public override string ToString() => "[object AMLDocument]";

    public string title
    {
        get => _origin.title;
        set => _origin.title = value;
    }
    public string? iconSrc
    {
        get => _origin.iconSrc;
        set => _origin.iconSrc = value;
    }
    public object body => _js_dispatcher.MarshalElement(_origin.body)!;
    public object createElement(string tag, object? options) =>
        _js_dispatcher.MarshalElement(_origin.createElement(tag, options));
    public object? getElementById(string id)
    {
        AML.Element? result = _origin.getElementById(id);
        if (result == null)
            return null;

        return _js_dispatcher.MarshalElement(result);
    }

    public void open(string url)
    {
        //TODO: open new document
        //_origin.
    }

    public void close()
    {
        //TODO: close this document
    }

    public void debug_stat()
    {
        Client.Client.RenderWriter.ConsolePrint(_origin.GetStatistics(""));
    }
}
