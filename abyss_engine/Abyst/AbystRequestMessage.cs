namespace AbyssCLI.Abyst;

public class AbystRequestMessage
{
    public AbystRequestMessage(HttpMethod method, string path)
    {
    }

    public override string ToString() => "abyst:local";
}
