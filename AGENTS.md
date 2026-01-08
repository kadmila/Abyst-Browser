# Agent Guidelines for Abyss Browser

This document provides essential information for AI coding agents working on the Abyss Browser project.

## Project Overview

Abyss Browser is a 3D browser engine built in C# (.NET 8.0) with peer-to-peer networking capabilities. The engine uses a custom markup language (AML - Abyss Markup Language) for rendering 3D content, integrates JavaScript via ClearScript V8, and communicates with the UI through Protocol Buffers.

**Working Directory**: `C:\workdir\github\Abyss-Browser\abyss_engine`

## Build Commands

### Full Build (Recommended)
```powershell
.\build_debug.ps1              # Builds ABI layer, updates timestamps, compiles C#
.\export_debug.ps1             # Exports to AbyssUI directories
```

### Direct .NET Build
```powershell
dotnet build AbyssCLI.csproj --configuration Debug
dotnet build AbyssCLI.csproj --configuration Release
dotnet clean                   # Clean build artifacts
dotnet restore                 # Restore NuGet packages
```

### ABI Layer Only
```powershell
cd ABI
.\build.ps1                    # Builds protobuf + code generation
.\build_protobuf.ps1           # Protobuf compilation only
```

### Running
```powershell
dotnet run
.\bin\Debug\net8.0\AbyssCLI.exe
```

## Testing

**No formal test framework is configured.** Tests are manual and located in `Test/` directory.

### Running Manual Tests
Tests must be called directly from code:
```csharp
ExternalDllTest.TestDllLoad();
ExternalDllTest.TestHostCreate();
ExternalDllTest.TestHostJoin();
ExternalDllTest.TestObjectSharing();
```

**When adding tests**: Create static test methods in `Test/` directory files and call them explicitly from `Program.cs` or a dedicated test runner.

## Code Style Guidelines

### Naming Conventions

**Classes**: PascalCase
```csharp
public class Element { }
public class Document { }
```

**Methods**:
- Public (JavaScript API compatible): camelCase
  ```csharp
  public void appendChild(Element child) { }
  public void setActive(bool active) { }
  ```
- Regular public/internal: PascalCase
  ```csharp
  public void WaitForEvent() { }
  public Task OpenWorld(string url) { }
  ```

**Fields**:
- Private: `_camelCase` with underscore prefix
  ```csharp
  private readonly Document _document;
  private static readonly BinaryReader _cin;
  ```
- Public readonly: PascalCase
  ```csharp
  public readonly int ElementId;
  public int RefCount;
  ```

**Constants**: PascalCase
```csharp
public const string BuildTime = "...";
```

### Imports and Namespaces

Use implicit usings (enabled in project). Always specify full namespaces:
```csharp
using AbyssCLI.ABI;
using AbyssCLI.Tool;
using AbyssCLI.Client;
using static AbyssCLI.ABI.UIAction.Types;  // For nested types
```

Namespace should match directory structure:
```csharp
namespace AbyssCLI.AML;
namespace AbyssCLI.Client;
```

### Types and Nullability

**Nullable reference types**: Disabled by default, enable selectively:
```csharp
#nullable enable
public Element? Parent;          // Nullable reference
public string tagName;           // Non-nullable (but not enforced)
#nullable disable
```

**Type preferences**:
- Use `var` for obvious types: `var client = new AbystClient();`
- Explicit types for clarity: `Dictionary<string, string> Attributes = [];`
- Collection expressions: `[]` instead of `new()`
  ```csharp
  public readonly List<Element> Children = [];
  public readonly Dictionary<string, string> Attributes = [];
  ```

### Formatting

**Braces**: Same-line for methods, properties; next-line acceptable for control flow
```csharp
public void Method() {
    if (condition)
    {
        // code
    }
}
```

**Lambda expressions**:
```csharp
public void CerrWriteLine(string message) => _cerr.WriteLine(message);
```

**String interpolation**: No specific preference; use what's clearest

### Architecture Patterns

**Partial Classes**: Split large classes by responsibility
```csharp
// Client.cs
public static partial class Client { }

// Client_Main.cs
public static partial class Client { }

// Client_UIActionHandlers.cs
public static partial class Client { }
```

**Primary Constructors** (C# 12): Use for simple initialization
```csharp
public class Cache(Action<HttpRequestMessage> http_requester, 
                   Action<AbystRequest> abyst_requester) { }
```

**IDisposable Pattern**: Always implement for managed resources
```csharp
public class Element : IDisposable
{
    public void Dispose()
    {
        // Cleanup logic
        GC.SuppressFinalize(this);
    }
    
    ~Element() => CerrWriteLine("fatal:::Element finalized without disposing");
}
```

**Memory Pressure**: Add for large unmanaged allocations
```csharp
GC.AddMemoryPressure(1_000_000_000);
```

### Error Handling

**Validation**: Throw exceptions for invalid states
```csharp
if (condition)
    throw new InvalidOperationException("Message");
    
if (argument == null)
    throw new ArgumentException("[null] is not AmlElement");
```

**Logging**: Use `Client.CerrWriteLine()` for diagnostics
```csharp
Client.CerrWriteLine("abyst server path: " + abyst_server_path);
Client.CerrWriteLine("fatal:::Element finalized without disposing");
```

**Native DLL errors**: Always check return values
```csharp
if (AbyssLib.Init() != 0)
    throw new Exception("failed to initialize abyssnet.dll");

if (!Host.IsValid())
{
    CerrWriteLine("host creation failed: " + AbyssLib.GetError().ToString());
    return;
}
```

### Unsafe Code and P/Invoke

**Allowed and common** for performance-critical paths:
```csharp
[DllImport("abyssnet.dll")]
static extern int GetVersion(byte* buf, int buflen);

unsafe {
    fixed (byte* pBytes = new byte[16])
    {
        // Pointer operations
    }
}
```

### Async/Await

**Heavily used** throughout the codebase:
```csharp
public static async Task Main() { }
public async Task<World> OpenWorld(string world_url) { }
```

Use `Task.Run()` for background work:
```csharp
Task.Run(async () => {
    HttpResponseMessage result = await HttpClient.SendAsync(request);
    // Process result
});
```

## Key Technologies

- **.NET 8.0**: Latest LTS version, use modern C# 12 features
- **Protocol Buffers**: For IPC between engine and UI renderer
- **ClearScript V8**: JavaScript engine integration
- **Native Interop**: `abyssnet.dll` for networking (P/Invoke heavy)
- **Async/Channel**: For message passing and async coordination

## Directory Structure

```
abyss_engine/
├── ABI/              # Application Binary Interface (protobuf definitions)
├── AML/              # Abyss Markup Language (DOM-like 3D elements)
├── Abyst/            # Abyst protocol client
├── Cache/            # Resource caching system
├── Client/           # Main client logic (partial classes)
├── HL/               # High-level abstractions
├── Test/             # Manual test classes
├── Tool/             # Utility classes (Error, Rc, URL parsing)
├── external_utils/   # Build utilities (Go code generator)
└── Program.cs        # Application entry point
```

## Important Notes

1. **No automatic code formatting** is configured. Follow existing style in similar files.
2. **Pragma warnings** are used to suppress specific IDE warnings:
   ```csharp
   #pragma warning disable IDE1006  // naming convention
   ```
3. **AllowUnsafeBlocks** is enabled project-wide; unsafe code is acceptable.
4. **ImplicitUsings** is enabled; common namespaces are auto-imported.
5. **Protocol Buffer changes** require running ABI build scripts to regenerate code.
6. **Native DLL** (`abyssnet.dll`) must be present in bin directory for runtime.

## When Making Changes

1. **Modifying .proto files**: Run `.\ABI\build.ps1` to regenerate C# classes
2. **Adding new features**: Update corresponding partial class files
3. **Native interop changes**: Update `AbyssLib.cs` and ensure DLL compatibility
4. **Testing changes**: Add manual test methods in `Test/` directory
5. **Build issues**: Clean with `dotnet clean` before rebuilding

## Common Patterns to Follow

- Use `Channel<T>` for async message queuing
- Reference count with `Rc<T>` for shared ownership
- Stream render actions via `RenderActionWriter`
- Read UI actions via `ReadProtoMessage()`
- Dispose elements properly to avoid memory leaks
- Use `SingleThreadTaskRunner` for single-threaded async work

---

**Last Updated**: 2026-01-09
