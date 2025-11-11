
#nullable enable
#pragma warning disable IDE1006 //naming convension
using Microsoft.ClearScript;
using Microsoft.ClearScript.JavaScript;
using Microsoft.ClearScript.V8;

namespace AbyssCLI.AML.JavaScriptAPI;

public class FetchApi
{
    public readonly V8ScriptEngine Engine;
    public FetchApi(V8ScriptEngine engine) => Engine = engine;

    // Fetch the content from a URL
    public object FetchAsync(object? url, object? options) =>
        JavaScriptExtensions.ToPromise(FetchInternalAsync(url as string ?? string.Empty, options as ScriptObject), Engine);
    private async Task<Response> FetchInternalAsync(string url, ScriptObject? options)
    {
        //Client.Client.RenderWriter.ConsolePrint("fetch called, option: " + options?.ToString());
        string method = options == null ? "GET" : options.GetProperty("method") as string ?? "GET";
        switch (method)
        {
        case "GET":
        {
            var response = await Client.Client.HttpClient.GetAsync(url);
            return new Response(this, response);
        }
        case "POST":
        {
            HttpContent content;
            var body_raw = options?.GetProperty("body");
            if (body_raw is string body)
            {
                content = new StringContent(body);
            }
            else
            {
                content = new StringContent("");
            }

            var response = await Client.Client.HttpClient.PostAsync(url, content);
            return new Response(this, response);
        }
        default:
            throw new Exception("unsupported http method");
        }
    }
}

public class Response
{
    private readonly FetchApi _origin;
    private readonly HttpResponseMessage _native_response;
    public readonly bool ok;
    public readonly int status;
    public readonly string statusText;
    internal Response(FetchApi origin, HttpResponseMessage native_response)
    {
        _origin = origin;
        _native_response = native_response;
        ok = native_response.IsSuccessStatusCode;
        status = (int)native_response.StatusCode;
        statusText = native_response.StatusCode.ToString();
    }
    public object text() => JavaScriptExtensions.ToPromise(_native_response.Content.ReadAsStringAsync(), _origin.Engine);
}