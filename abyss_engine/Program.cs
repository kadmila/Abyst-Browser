using AbyssCLI.Client;
using AbyssCLI;
using System.IO;

internal class Program
{
    public static async Task Main()
    {
        await Test();

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

    public static async Task Test()
    {
        try
        {
            // 1) Initialize library
            _ = AbyssLibB.Initialize();

            // 2) Create host from testkey.pem
            string pemPath = "../../../testkey.pem";
            if (!File.Exists(pemPath))
            {
                Client.CerrWriteLine($"Error: {pemPath} not found in project root directory");
                Environment.Exit(1);
            }

            byte[] keyBytes = File.ReadAllBytes(pemPath);

            var (host, hostError) = AbyssLibB.Host.Create(keyBytes);
            if (hostError != null)
            {
                Client.CerrWriteLine($"Error creating host: {hostError.Message}");
                Environment.Exit(1);
            }

            if (host == null)
            {
                Client.CerrWriteLine("Error: Failed to create host (unknown error)");
                Environment.Exit(1);
            }

            using (host)
            {
                Client.CerrWriteLine($"Host created successfully with ID: {host.ID}");

                var bind_error = host.Bind();
                if (bind_error != null)
                {
                    Client.CerrWriteLine(bind_error.Message);
                    Environment.Exit(1);
                }

                host.Serve();

                using (var abystClient = host.NewAbystClient())
                {
                    var (response, error) = await abystClient.Get(host.ID, "path");
                    if (error != null)
                    {
                        Console.WriteLine(error.Message);
                    }
                    else
                    {
                        Console.WriteLine(response.GetAllHeaders());
                    }
                }

                // 3) Create CollocatedHttp3Client from host
                using (var http3Client = host.NewCollocatedHttp3Client())
                {
                    var (response, error) = await http3Client.Get("https://localhost:4433");
                    
                    if (error != null)
                    {
                        Client.CerrWriteLine($"HTTP request failed: {error.Message}");
                        Environment.Exit(1);
                    }

                    if (response == null)
                    {
                        Client.CerrWriteLine("HTTP request returned null response");
                        Environment.Exit(1);
                    }

                    using (response)
                    {
                        Client.CerrWriteLine($"HTTP Response Status: {response.StatusCode}");
                        
                        // Log headers
                        string allHeaders = response.GetAllHeaders();
                        if (!string.IsNullOrEmpty(allHeaders))
                        {
                            Client.CerrWriteLine("Response Headers:");
                            Client.CerrWriteLine(allHeaders);
                        }
                        
                        // Log body
                        byte[] bodyBytes = response.ReadAllBody();
                        if (bodyBytes.Length > 0)
                        {
                            string bodyText = System.Text.Encoding.UTF8.GetString(bodyBytes);
                            Client.CerrWriteLine($"Response Body ({bodyBytes.Length} bytes):");
                            Client.CerrWriteLine(bodyText);
                        }
                        else
                        {
                            Client.CerrWriteLine("Response Body: (empty)");
                        }

                        Client.CerrWriteLine("HTTP request completed successfully");
                    }
                }
            }
        }
        catch (FileNotFoundException ex)
        {
            Client.CerrWriteLine($"File not found: {ex.Message}");
            Environment.Exit(1);
        }
        catch (IOException ex)
        {
            Client.CerrWriteLine($"IO error: {ex.Message}");
            Environment.Exit(1);
        }
        catch (System.Exception ex)
        {
            Client.CerrWriteLine($"Unexpected error: {ex.GetType().Name}: {ex.Message}");
            Client.CerrWriteLine($"Stack trace: {ex.StackTrace}");
            Environment.Exit(1);
        }

        Environment.Exit(0);
    }
}
