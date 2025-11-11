namespace AbyssCLI.Tool;
public class AbyssURL
{
    public string Raw
    {
        get; set;
    }
    public string Scheme
    {
        get; set;
    }
    public string Id { get; set; } = ""; // For abyss/abyst
    public List<(string Ip, int Port)> AddressCandidates { get; set; } = [];
    public string Path { get; set; } = ""; // For abyss/abyst
    public Uri StandardUri
    {
        get; set;
    } // For standard and abyst URIs

    public override string ToString() => Raw;
}

public static class AbyssURLParser
{
    public static bool TryParse(string _input, out AbyssURL result)
    {
        string input = _input.Trim();
        if (input.StartsWith("abyss:"))
        {
            return TryParseAbyss(input, out result);
        }
        else if (input.StartsWith("abyst:"))
        {
            return TryParseAbyst(input, out result);
        }
        else
        {
            try
            {
                var parsed_uri = new Uri(input);
                result = new AbyssURL
                {
                    Raw = input,
                    Scheme = parsed_uri.Scheme,
                    StandardUri = parsed_uri,
                };
                return true;
            }
            catch
            {
                result = new AbyssURL();
                return false;
            }
        }
    }
    private static string CalculateRelativePath(string basePath, string targetPath)
    {
        // Step 1: Check if the target path starts with "/"
        if (targetPath.StartsWith('/'))
            return targetPath;

        // Step 2: Check if the target path starts with "./"
        if (targetPath.StartsWith("./"))
            return CalculateRelativePath(basePath, targetPath[2..]);

        // Step 3: Check if the target path starts with "../"
        if (targetPath.StartsWith("../"))
        {
            // Remove the last part of the base path
            int lastSlashIndex = basePath.LastIndexOf('/');
            if (lastSlashIndex == 0)
                throw new InvalidOperationException("Base path has no parent directory.");

            return CalculateRelativePath(basePath[..lastSlashIndex], targetPath[3..]);
        }

        return basePath[..(basePath.LastIndexOf('/') + 1)] + targetPath;
    }
    public static bool TryParseFrom(string _input, AbyssURL origin, out AbyssURL result)
    {
        string input = _input.Trim();
        if (!input.Contains(':'))
        {
            if (origin.Scheme == "abyss")
            {
                return AbyssURLParser.TryParse("abyst:" + origin.Id + "/" + input.Trim().TrimStart('/'), out result);
            }
            else if (origin.Scheme == "abyst")
            {
                return AbyssURLParser.TryParse("abyst:" + origin.Id + CalculateRelativePath('/' + origin.Path, input), out result);
            }

            //web address
            try
            {
                var parsed_uri = new Uri(origin.StandardUri, input);
                result = new AbyssURL
                {
                    Raw = parsed_uri.ToString(),
                    Scheme = parsed_uri.Scheme,
                    StandardUri = parsed_uri,
                };
                return true;
            }
            catch
            {
                result = default;
                return false;
            }
        }

        return AbyssURLParser.TryParse(input, out result);
    }
    private const string Base58Chars = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz";
    private static bool IsValidPeerID(string input)
    {
        if (string.IsNullOrEmpty(input) || input.Length < 32 || !char.IsUpper(input[0])) //id version code
        {
            return false;
        }

        foreach (char c in input[1..])
        {
            if (!Base58Chars.Contains(c))
                return false;
        }

        return true;
    }

    private static bool TryParseAbyss(string input, out AbyssURL result)
    {
        result = new AbyssURL
        {
            Raw = input,
            Scheme = "abyss"
        };

        string body = input["abyss:".Length..];
        if (string.IsNullOrEmpty(body))
        {
            return false;
        }

        int addr_start_pos = body.IndexOf(':');
        int path_start_pos = body.IndexOf('/');
        if (addr_start_pos == -1) //there is no address section.
        {

            if (path_start_pos == -1)
            {
                //there is no path either, body is the id.
                //check if body is a valid id.
                if (!IsValidPeerID(body)) //id version code
                {
                    return false;
                }
                result.Id = body;
                return true;
            }

            //only path.
            string _peer_id = body[..path_start_pos];
            if (!IsValidPeerID(_peer_id)) //id version code
            {
                return false;
            }
            result.Id = _peer_id;
            result.Path = body[(path_start_pos + 1)..];
            return true;
        }

        //first, detach id that comes before addresses.
        string peer_id = body[..addr_start_pos];
        if (!IsValidPeerID(peer_id))
        {
            return false;
        }
        result.Id = peer_id;
        //now, it is also certain that path starts after the addresses, as the peer ID cannot contain '/'.

        string addr_part = path_start_pos != -1 ? body[(addr_start_pos + 1)..path_start_pos] : body[(addr_start_pos + 1)..];
        result.Path = path_start_pos != -1 ? body[(path_start_pos + 1)..] : "";

        //// Parse IP:Port list (IPv4 only)
        //foreach (var ep in addr_part.Split('|', StringSplitOptions.RemoveEmptyEntries))
        //{
        //    var parts = ep.Split(':');
        //    if (parts.Length == 2 &&
        //        System.Net.IPAddress.TryParse(parts[0], out _) &&
        //        int.TryParse(parts[1], out int port))
        //    {
        //        result.AddressCandidates.Add((parts[0], port));
        //    }
        //}

        // Parse IP:Port list
        foreach (string ep in addr_part.Split('|', StringSplitOptions.RemoveEmptyEntries))
        {
            string ipPart;
            string portPart;

            // If the endpoint starts with '[', we treat it as a bracketed IPv6 address
            if (ep.StartsWith('['))
            {
                int closeBracketIndex = ep.IndexOf(']');
                // If no closing bracket is found or it's the only character, skip
                if (closeBracketIndex <= 0)
                    continue;

                // Extract the IP portion from inside the brackets
                ipPart = ep[1..closeBracketIndex];

                // If there is a colon after ']', treat what's after it as the port
                if (closeBracketIndex + 1 < ep.Length && ep[closeBracketIndex + 1] == ':')
                {
                    portPart = ep[(closeBracketIndex + 2)..];
                }
                else
                {
                    // No port specified
                    continue;
                }
            }
            else
            {
                // For IPv4, domain names, or any non-bracketed input, split on colon
                string[] parts2 = ep.Split(':');
                if (parts2.Length == 2)
                {
                    ipPart = parts2[0];
                    portPart = parts2[1];
                }
                else
                {
                    // Possibly invalid or missing port
                    continue;
                }
            }

            if (System.Net.IPAddress.TryParse(ipPart, out _) && int.TryParse(portPart, out int port))
            {
                result.AddressCandidates.Add((ipPart, port));
            }
        }

        return true;
    }

    private static bool TryParseAbyst(string input, out AbyssURL result)
    {
        result = new AbyssURL
        {
            Raw = input,
            Scheme = "abyst"
        };
        string body = input["abyst:".Length..];

        // Extract ID and path/query using first '/'
        int slashIndex = body.IndexOf('/');
        if (slashIndex == -1)
        {
            if (!IsValidPeerID(body))
            {
                return false;
            }
            result.Id = body;
            return true;
        }

        result.Id = body[..slashIndex];
        result.Path = body[(slashIndex + 1)..];
        return true;
    }
}
