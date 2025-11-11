using AbyssCLI.Client;
internal class Program
{
    public static async Task Main()
    {
        try
        {
            Client.Init();
            await Client.Run();
            Client.CerrWriteLine("AbyssCLI terminated peacefully");
        }
        catch (Exception ex)
        {
            Client.CerrWriteLine("***FATAL::ABYSS_CLI TERMINATED***\n" + ex.ToString());
        }
    }
}
