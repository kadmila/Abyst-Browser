using System.Runtime.CompilerServices;
using System.Runtime.InteropServices;
using System.Text;

#nullable enable
namespace AbyssCLI;

/// <summary>
/// New version of AbyssLib with improved API based on abyssnet.h header.
/// </summary>
public static class AbyssLibB
{
    private const string DllName = "abyssnet.dll";

    #region P/Invoke Declarations

    [DllImport(DllName)] private static extern int Init();
    [DllImport(DllName)] private static extern int GetErrorBodyLength(IntPtr h_error);
    [DllImport(DllName)] private static extern unsafe int GetErrorBody(IntPtr h_error, byte* buf_ptr, int buf_len);
    [DllImport(DllName)] private static extern void CloseError(IntPtr h_error);

    [DllImport(DllName)] private static extern unsafe IntPtr NewHost(byte* root_key_ptr, int root_key_len, IntPtr* host_out);
    [DllImport(DllName)] private static extern void CloseHost(IntPtr h);
    [DllImport(DllName)] private static extern void Host_Run(IntPtr h);
    [DllImport(DllName)] private static extern unsafe IntPtr Host_WaitForEvent(IntPtr h, int* event_type_out, IntPtr* event_handle_out);
    [DllImport(DllName)] private static extern void CloseEvent(IntPtr h);
    [DllImport(DllName)] private static extern unsafe IntPtr Host_OpenWorld(IntPtr h, byte* world_url_ptr, int world_url_len, IntPtr* world_handle_out);
    [DllImport(DllName)] private static extern unsafe IntPtr Host_JoinWorld(IntPtr h, IntPtr h_peer, byte* path_ptr, int path_len, IntPtr* world_handle_out);
    [DllImport(DllName)] private static extern void CloseWorld(IntPtr h_world);
    [DllImport(DllName)] private static extern unsafe IntPtr Host_ExposeWorldForJoin(IntPtr h, IntPtr h_world, byte* path_ptr, int path_len);
    [DllImport(DllName)] private static extern void Host_HideWorld(IntPtr h, IntPtr h_world);
    [DllImport(DllName)] private static extern unsafe int Host_LocalAddrCandidates(IntPtr h, byte* buf_ptr, int buf_len);
    [DllImport(DllName)] private static extern unsafe int Host_ID(IntPtr h, byte* buf_ptr, int buf_len);
    [DllImport(DllName)] private static extern unsafe int Host_RootCertificate(IntPtr h, byte* buf_ptr, int buf_len);
    [DllImport(DllName)] private static extern unsafe int Host_HandshakeKeyCertificate(IntPtr h, byte* buf_ptr, int buf_len);
    [DllImport(DllName)] private static extern unsafe IntPtr Host_AppendKnownPeer(IntPtr h, byte* root_cert_ptr, int root_cert_len, byte* handshake_info_cert_ptr, int handshake_info_cert_len);
    [DllImport(DllName)] private static extern unsafe IntPtr Host_EraseKnownPeer(IntPtr h, byte* id_ptr, int id_len);
    [DllImport(DllName)] private static extern unsafe IntPtr Host_Dial(IntPtr h, byte* id_ptr, int id_len);
    [DllImport(DllName)] private static extern unsafe IntPtr Host_ConfigAbystGateway(IntPtr h, byte* config_ptr, int config_len);
    [DllImport(DllName)] private static extern IntPtr Host_NewAbystClient(IntPtr h);
    [DllImport(DllName)] private static extern void CloseAbyssClient(IntPtr h);
    [DllImport(DllName)] private static extern IntPtr Host_NewCollocatedHttp3Client(IntPtr h);
    [DllImport(DllName)] private static extern void CloseAbyssClientCollocatedHttp3Client(IntPtr h);

    [DllImport(DllName)] private static extern void ClosePeer(IntPtr h_peer);

    [DllImport(DllName)] private static extern unsafe IntPtr AbystClient_Get(IntPtr h_client, IntPtr h_event, byte* peer_id_ptr, int peer_id_len, byte* path_ptr, int path_len, IntPtr* result_handle_out);
    [DllImport(DllName)] private static extern unsafe IntPtr AbystClient_Post(IntPtr h_client, IntPtr h_event, byte* peer_id_ptr, int peer_id_len, byte* path_ptr, int path_len, byte* content_type_ptr, int content_type_len, byte* body_ptr, int body_len, IntPtr* result_handle_out);
    [DllImport(DllName)] private static extern unsafe IntPtr AbystClient_Head(IntPtr h_client, IntPtr h_event, byte* peer_id_ptr, int peer_id_len, byte* path_ptr, int path_len, IntPtr* result_handle_out);

    [DllImport(DllName)] private static extern unsafe IntPtr Http3Client_Get(IntPtr h_client, IntPtr h_event, byte* url_ptr, int url_len, IntPtr* result_handle_out);
    [DllImport(DllName)] private static extern unsafe IntPtr Http3Client_Post(IntPtr h_client, IntPtr h_event, byte* url_ptr, int url_len, byte* content_type_ptr, int content_type_len, byte* body_ptr, int body_len, IntPtr* result_handle_out);
    [DllImport(DllName)] private static extern unsafe IntPtr Http3Client_Head(IntPtr h_client, IntPtr h_event, byte* url_ptr, int url_len, IntPtr* result_handle_out);

    [DllImport(DllName)] private static extern void CloseHttpIOResult(IntPtr h_result);
    [DllImport(DllName)] private static extern unsafe IntPtr HttpIOResult_Unpack(IntPtr h_result, IntPtr* response_handle_out);
    [DllImport(DllName)] private static extern int HttpResponse_StatusCode(IntPtr h_response);
    [DllImport(DllName)] private static extern unsafe int HttpResponse_GetHeader(IntPtr h_response, byte* key_ptr, int key_len, byte* value_buf_ptr, int value_buf_len);
    [DllImport(DllName)] private static extern unsafe int HttpResponse_GetAllHeaders(IntPtr h_response, byte* buf_ptr, int buf_len);
    [DllImport(DllName)] private static extern unsafe int HttpResponse_ReadBody(IntPtr h_response, byte* buf_ptr, int buf_len);
    [DllImport(DllName)] private static extern void CloseHttpResponse(IntPtr h_response);

    [DllImport(DllName)] private static extern unsafe void World_AcceptSession(IntPtr h_host, IntPtr h_world, IntPtr h_peer, byte* peer_session_id_buf);
    [DllImport(DllName)] private static extern unsafe void World_DeclineSession(IntPtr h_host, IntPtr h_world, IntPtr h_peer, byte* peer_session_id_buf, int code, byte* message_buf_ptr, int message_buf_len);
    [DllImport(DllName)] private static extern void World_Close(IntPtr h_host, IntPtr h_world);
    [DllImport(DllName)] private static extern unsafe void World_ObjectAppend(IntPtr h_host, IntPtr h_world, int peer_count, IntPtr* h_peers, byte** peer_session_id_bufs, int object_count, byte** object_id_bufs, float** object_transform_bufs, byte** object_addr_bufs, int object_addr_buf_len);
    [DllImport(DllName)] private static extern unsafe void World_ObjectDelete(IntPtr h_host, IntPtr h_world, int peer_count, IntPtr* h_peers, byte** peer_session_id_bufs, int object_count, byte** object_id_bufs);

    // Event query functions
    [DllImport(DllName)] private static extern unsafe int Event_WorldEnter_Query(IntPtr h_event, byte* world_session_id_buf, byte* url_buf_ptr, int url_buf_len);
    [DllImport(DllName)] private static extern unsafe int Event_SessionRequest_Query(IntPtr h_event, byte* world_session_id_buf, byte* peer_session_id_buf, byte* peer_id_buf_ptr, int peer_id_buf_len);
    [DllImport(DllName)] private static extern unsafe int Event_SessionReady_Query(IntPtr h_event, byte* world_session_id_buf, byte* peer_session_id_buf, byte* peer_id_buf_ptr, int peer_id_buf_len);
    [DllImport(DllName)] private static extern unsafe int Event_SessionClose_Query(IntPtr h_event, byte* world_session_id_buf, byte* peer_session_id_buf, byte* peer_id_buf_ptr, int peer_id_buf_len);
    [DllImport(DllName)] private static extern unsafe int Event_ObjectAppend_Query(IntPtr h_event, byte* world_session_id_buf, byte* peer_session_id_buf, byte* peer_id_buf_ptr, int peer_id_buf_len, int* object_count_out);
    [DllImport(DllName)] private static extern unsafe int Event_ObjectAppend_GetObjects(IntPtr h_event, byte** object_id_bufs, float** object_transform_bufs, byte** object_addr_bufs, int object_addr_buf_len);
    [DllImport(DllName)] private static extern unsafe int Event_ObjectDelete_Query(IntPtr h_event, byte* world_session_id_buf, byte* peer_session_id_buf, byte* peer_id_buf_ptr, int peer_id_buf_len, int* object_count_out);
    [DllImport(DllName)] private static extern unsafe int Event_ObjectDelete_GetObjectIDs(IntPtr h_event, byte** object_id_bufs);
    [DllImport(DllName)] private static extern unsafe int Event_WorldLeave_Query(IntPtr h_event, byte* world_session_id_buf, int* code_out, byte* message_buf_ptr, int message_buf_len);
    [DllImport(DllName)] private static extern unsafe int Event_PeerConnected_Query(IntPtr h_event, IntPtr* peer_handle_out, byte* peer_id_buf_ptr, int peer_id_buf_len);
    [DllImport(DllName)] private static extern unsafe int Event_PeerDisconnected_Query(IntPtr h_event, byte* peer_id_buf_ptr, int peer_id_buf_len);

    #endregion

    #region Initialization

    public static int Initialize() => Init();

    #endregion

    #region Error Handling

    public class Error
    {
        public string Message { get; }

        public Error(IntPtr handle)
        {
            if (handle == IntPtr.Zero)
            {
                Message = string.Empty;
                return;
            }

            unsafe
            {
                int msgLen = GetErrorBodyLength(handle);
                byte[] buf = new byte[msgLen];
                fixed (byte* bufPtr = buf)
                {
                    int len = GetErrorBody(handle, bufPtr, buf.Length);
                    CloseError(handle);

                    if (len != buf.Length)
                    {
                        Message = "Error: fatal DLL corruption: failed to get error body";
                        return;
                    }
                    Message = Encoding.UTF8.GetString(buf);
                }
            }
        }
    }

    #endregion

    #region Host

    public class Host : IDisposable
    {
        private IntPtr _handle;

        public bool IsValid => _handle != IntPtr.Zero;

        private Host(IntPtr handle)
        {
            _handle = handle;
        }

        public static (Host?, Error?) Create(byte[] rootKey)
        {
            unsafe
            {
                fixed (byte* keyPtr = rootKey)
                {
                    IntPtr hostHandle;
                    IntPtr errHandle = NewHost(keyPtr, rootKey.Length, &hostHandle);
                    if (errHandle != IntPtr.Zero)
                        return (null, new Error(errHandle));
                    return (new Host(hostHandle), null);
                }
            }
        }

        public void Run() => Host_Run(_handle);

        public (EventType Type, Event? Event, Error?) WaitForEvent()
        {
            unsafe
            {
                int eventType;
                IntPtr eventHandle;
                IntPtr errHandle = Host_WaitForEvent(_handle, &eventType, &eventHandle);
                if (errHandle != IntPtr.Zero)
                    return (EventType.None, null, new Error(errHandle));
                return ((EventType)eventType, new Event(eventHandle, (EventType)eventType), null);
            }
        }

        public (World?, Error?) OpenWorld(string worldUrl)
        {
            byte[] urlBytes = Encoding.UTF8.GetBytes(worldUrl);
            unsafe
            {
                fixed (byte* urlPtr = urlBytes)
                {
                    IntPtr worldHandle;
                    IntPtr errHandle = Host_OpenWorld(_handle, urlPtr, urlBytes.Length, &worldHandle);
                    if (errHandle != IntPtr.Zero)
                        return (null, new Error(errHandle));
                    return (new World(worldHandle, this), null);
                }
            }
        }

        public (World?, Error?) JoinWorld(Peer peer, string path)
        {
            byte[] pathBytes = Encoding.UTF8.GetBytes(path);
            unsafe
            {
                fixed (byte* pathPtr = pathBytes)
                {
                    IntPtr worldHandle;
                    IntPtr errHandle = Host_JoinWorld(_handle, peer.Handle, pathPtr, pathBytes.Length, &worldHandle);
                    if (errHandle != IntPtr.Zero)
                        return (null, new Error(errHandle));
                    return (new World(worldHandle, this), null);
                }
            }
        }

        public Error? ExposeWorldForJoin(World world, string path)
        {
            byte[] pathBytes = Encoding.UTF8.GetBytes(path);
            unsafe
            {
                fixed (byte* pathPtr = pathBytes)
                {
                    IntPtr errHandle = Host_ExposeWorldForJoin(_handle, world.Handle, pathPtr, pathBytes.Length);
                    if (errHandle != IntPtr.Zero)
                        return new Error(errHandle);
                    return null;
                }
            }
        }

        public void HideWorld(World world) => Host_HideWorld(_handle, world.Handle);

        public string GetLocalAddrCandidates()
        {
            unsafe
            {
                byte[] buf = new byte[4096];
                fixed (byte* bufPtr = buf)
                {
                    int len = Host_LocalAddrCandidates(_handle, bufPtr, buf.Length);
                    return len > 0 ? Encoding.UTF8.GetString(buf, 0, len) : string.Empty;
                }
            }
        }

        public string GetID()
        {
            unsafe
            {
                byte[] buf = new byte[256];
                fixed (byte* bufPtr = buf)
                {
                    int len = Host_ID(_handle, bufPtr, buf.Length);
                    return len > 0 ? Encoding.UTF8.GetString(buf, 0, len) : string.Empty;
                }
            }
        }

        public byte[] GetRootCertificate()
        {
            unsafe
            {
                byte[] buf = new byte[4096];
                fixed (byte* bufPtr = buf)
                {
                    int len = Host_RootCertificate(_handle, bufPtr, buf.Length);
                    if (len <= 0) return [];
                    byte[] result = new byte[len];
                    Array.Copy(buf, result, len);
                    return result;
                }
            }
        }

        public byte[] GetHandshakeKeyCertificate()
        {
            unsafe
            {
                byte[] buf = new byte[4096];
                fixed (byte* bufPtr = buf)
                {
                    int len = Host_HandshakeKeyCertificate(_handle, bufPtr, buf.Length);
                    if (len <= 0) return [];
                    byte[] result = new byte[len];
                    Array.Copy(buf, result, len);
                    return result;
                }
            }
        }

        public Error? AppendKnownPeer(byte[] rootCert, byte[] handshakeInfoCert)
        {
            unsafe
            {
                fixed (byte* rootPtr = rootCert)
                fixed (byte* hsPtr = handshakeInfoCert)
                {
                    IntPtr errHandle = Host_AppendKnownPeer(_handle, rootPtr, rootCert.Length, hsPtr, handshakeInfoCert.Length);
                    if (errHandle != IntPtr.Zero)
                        return new Error(errHandle);
                    return null;
                }
            }
        }

        public Error? EraseKnownPeer(string id)
        {
            byte[] idBytes = Encoding.UTF8.GetBytes(id);
            unsafe
            {
                fixed (byte* idPtr = idBytes)
                {
                    IntPtr errHandle = Host_EraseKnownPeer(_handle, idPtr, idBytes.Length);
                    if (errHandle != IntPtr.Zero)
                        return new Error(errHandle);
                    return null;
                }
            }
        }

        public Error? Dial(string id)
        {
            byte[] idBytes = Encoding.UTF8.GetBytes(id);
            unsafe
            {
                fixed (byte* idPtr = idBytes)
                {
                    IntPtr errHandle = Host_Dial(_handle, idPtr, idBytes.Length);
                    if (errHandle != IntPtr.Zero)
                        return new Error(errHandle);
                    return null;
                }
            }
        }

        public Error? ConfigAbystGateway(string config)
        {
            byte[] configBytes = Encoding.UTF8.GetBytes(config);
            unsafe
            {
                fixed (byte* configPtr = configBytes)
                {
                    IntPtr errHandle = Host_ConfigAbystGateway(_handle, configPtr, configBytes.Length);
                    if (errHandle != IntPtr.Zero)
                        return new Error(errHandle);
                    return null;
                }
            }
        }

        public AbystClient NewAbystClient()
        {
            IntPtr clientHandle = Host_NewAbystClient(_handle);
            return new AbystClient(clientHandle);
        }

        public Http3Client NewCollocatedHttp3Client()
        {
            IntPtr clientHandle = Host_NewCollocatedHttp3Client(_handle);
            return new Http3Client(clientHandle);
        }

        internal IntPtr Handle => _handle;

        public void Dispose()
        {
            if (_handle != IntPtr.Zero)
            {
                CloseHost(_handle);
                _handle = IntPtr.Zero;
            }
            GC.SuppressFinalize(this);
        }

        ~Host() => Dispose();
    }

    #endregion

    #region Event Types

    public enum EventType
    {
        None = 0,
        WorldEnter = 1,
        SessionRequest = 2,
        SessionReady = 3,
        SessionClose = 4,
        ObjectAppend = 5,
        ObjectDelete = 6,
        WorldLeave = 7,
        PeerConnected = 8,
        PeerDisconnected = 9,
    }

    public class Event : IDisposable
    {
        private IntPtr _handle;
        private bool _disposed;
        public EventType Type { get; }

        internal Event(IntPtr handle, EventType type)
        {
            _handle = handle;
            Type = type;
        }

        public WorldEnterEventData? QueryWorldEnter()
        {
            if (Type != EventType.WorldEnter) return null;
            unsafe
            {
                byte[] worldSessionId = new byte[16];
                byte[] urlBuf = new byte[2048];
                fixed (byte* wsidPtr = worldSessionId)
                fixed (byte* urlPtr = urlBuf)
                {
                    int urlLen = Event_WorldEnter_Query(_handle, wsidPtr, urlPtr, urlBuf.Length);
                    if (urlLen < 0) return null;
                    return new WorldEnterEventData
                    {
                        WorldSessionId = worldSessionId,
                        Url = Encoding.UTF8.GetString(urlBuf, 0, urlLen)
                    };
                }
            }
        }

        public SessionEventData? QuerySessionRequest()
        {
            if (Type != EventType.SessionRequest) return null;
            unsafe
            {
                byte[] worldSessionId = new byte[16];
                byte[] peerSessionId = new byte[16];
                byte[] peerIdBuf = new byte[256];
                fixed (byte* wsidPtr = worldSessionId)
                fixed (byte* psidPtr = peerSessionId)
                fixed (byte* pidPtr = peerIdBuf)
                {
                    int peerIdLen = Event_SessionRequest_Query(_handle, wsidPtr, psidPtr, pidPtr, peerIdBuf.Length);
                    if (peerIdLen < 0) return null;
                    return new SessionEventData
                    {
                        WorldSessionId = worldSessionId,
                        PeerSessionId = peerSessionId,
                        PeerId = Encoding.UTF8.GetString(peerIdBuf, 0, peerIdLen)
                    };
                }
            }
        }

        public SessionEventData? QuerySessionReady()
        {
            if (Type != EventType.SessionReady) return null;
            unsafe
            {
                byte[] worldSessionId = new byte[16];
                byte[] peerSessionId = new byte[16];
                byte[] peerIdBuf = new byte[256];
                fixed (byte* wsidPtr = worldSessionId)
                fixed (byte* psidPtr = peerSessionId)
                fixed (byte* pidPtr = peerIdBuf)
                {
                    int peerIdLen = Event_SessionReady_Query(_handle, wsidPtr, psidPtr, pidPtr, peerIdBuf.Length);
                    if (peerIdLen < 0) return null;
                    return new SessionEventData
                    {
                        WorldSessionId = worldSessionId,
                        PeerSessionId = peerSessionId,
                        PeerId = Encoding.UTF8.GetString(peerIdBuf, 0, peerIdLen)
                    };
                }
            }
        }

        public SessionEventData? QuerySessionClose()
        {
            if (Type != EventType.SessionClose) return null;
            unsafe
            {
                byte[] worldSessionId = new byte[16];
                byte[] peerSessionId = new byte[16];
                byte[] peerIdBuf = new byte[256];
                fixed (byte* wsidPtr = worldSessionId)
                fixed (byte* psidPtr = peerSessionId)
                fixed (byte* pidPtr = peerIdBuf)
                {
                    int peerIdLen = Event_SessionClose_Query(_handle, wsidPtr, psidPtr, pidPtr, peerIdBuf.Length);
                    if (peerIdLen < 0) return null;
                    return new SessionEventData
                    {
                        WorldSessionId = worldSessionId,
                        PeerSessionId = peerSessionId,
                        PeerId = Encoding.UTF8.GetString(peerIdBuf, 0, peerIdLen)
                    };
                }
            }
        }

        public ObjectAppendEventData? QueryObjectAppend()
        {
            if (Type != EventType.ObjectAppend) return null;
            unsafe
            {
                byte[] worldSessionId = new byte[16];
                byte[] peerSessionId = new byte[16];
                byte[] peerIdBuf = new byte[256];
                int objectCount;
                fixed (byte* wsidPtr = worldSessionId)
                fixed (byte* psidPtr = peerSessionId)
                fixed (byte* pidPtr = peerIdBuf)
                {
                    int peerIdLen = Event_ObjectAppend_Query(_handle, wsidPtr, psidPtr, pidPtr, peerIdBuf.Length, &objectCount);
                    if (peerIdLen < 0) return null;

                    var data = new ObjectAppendEventData
                    {
                        WorldSessionId = worldSessionId,
                        PeerSessionId = peerSessionId,
                        PeerId = Encoding.UTF8.GetString(peerIdBuf, 0, peerIdLen),
                        Objects = new ObjectInfo[objectCount]
                    };

                    if (objectCount > 0)
                    {
                        // Allocate buffers for object data
                        const int ObjectIdSize = 16;
                        const int TransformSize = 16; // 16 floats
                        const int AddrBufLen = 256;

                        byte[][] objectIdBuffers = new byte[objectCount][];
                        float[][] transformBuffers = new float[objectCount][];
                        byte[][] addrBuffers = new byte[objectCount][];

                        for (int i = 0; i < objectCount; i++)
                        {
                            objectIdBuffers[i] = new byte[ObjectIdSize];
                            transformBuffers[i] = new float[TransformSize];
                            addrBuffers[i] = new byte[AddrBufLen];
                        }

                        fixed (byte* id0 = objectIdBuffers[0])
                        fixed (float* tr0 = transformBuffers[0])
                        fixed (byte* addr0 = addrBuffers[0])
                        {
                            byte*[] idPtrs = new byte*[objectCount];
                            float*[] trPtrs = new float*[objectCount];
                            byte*[] addrPtrs = new byte*[objectCount];

                            for (int i = 0; i < objectCount; i++)
                            {
                                fixed (byte* idp = objectIdBuffers[i])
                                fixed (float* trp = transformBuffers[i])
                                fixed (byte* addrp = addrBuffers[i])
                                {
                                    idPtrs[i] = idp;
                                    trPtrs[i] = trp;
                                    addrPtrs[i] = addrp;
                                }
                            }

                            fixed (byte** idPtrsPtr = idPtrs)
                            fixed (float** trPtrsPtr = trPtrs)
                            fixed (byte** addrPtrsPtr = addrPtrs)
                            {
                                Event_ObjectAppend_GetObjects(_handle, idPtrsPtr, trPtrsPtr, addrPtrsPtr, AddrBufLen);
                            }
                        }

                        for (int i = 0; i < objectCount; i++)
                        {
                            data.Objects[i] = new ObjectInfo
                            {
                                Id = objectIdBuffers[i],
                                Transform = transformBuffers[i],
                                Address = Encoding.UTF8.GetString(addrBuffers[i]).TrimEnd('\0')
                            };
                        }
                    }

                    return data;
                }
            }
        }

        public ObjectDeleteEventData? QueryObjectDelete()
        {
            if (Type != EventType.ObjectDelete) return null;
            unsafe
            {
                byte[] worldSessionId = new byte[16];
                byte[] peerSessionId = new byte[16];
                byte[] peerIdBuf = new byte[256];
                int objectCount;
                fixed (byte* wsidPtr = worldSessionId)
                fixed (byte* psidPtr = peerSessionId)
                fixed (byte* pidPtr = peerIdBuf)
                {
                    int peerIdLen = Event_ObjectDelete_Query(_handle, wsidPtr, psidPtr, pidPtr, peerIdBuf.Length, &objectCount);
                    if (peerIdLen < 0) return null;

                    var data = new ObjectDeleteEventData
                    {
                        WorldSessionId = worldSessionId,
                        PeerSessionId = peerSessionId,
                        PeerId = Encoding.UTF8.GetString(peerIdBuf, 0, peerIdLen),
                        ObjectIds = new byte[objectCount][]
                    };

                    if (objectCount > 0)
                    {
                        const int ObjectIdSize = 16;
                        byte[][] objectIdBuffers = new byte[objectCount][];
                        for (int i = 0; i < objectCount; i++)
                            objectIdBuffers[i] = new byte[ObjectIdSize];

                        fixed (byte* id0 = objectIdBuffers[0])
                        {
                            byte*[] idPtrs = new byte*[objectCount];
                            for (int i = 0; i < objectCount; i++)
                            {
                                fixed (byte* idp = objectIdBuffers[i])
                                    idPtrs[i] = idp;
                            }

                            fixed (byte** idPtrsPtr = idPtrs)
                            {
                                Event_ObjectDelete_GetObjectIDs(_handle, idPtrsPtr);
                            }
                        }

                        data.ObjectIds = objectIdBuffers;
                    }

                    return data;
                }
            }
        }

        public WorldLeaveEventData? QueryWorldLeave()
        {
            if (Type != EventType.WorldLeave) return null;
            unsafe
            {
                byte[] worldSessionId = new byte[16];
                byte[] messageBuf = new byte[1024];
                int code;
                fixed (byte* wsidPtr = worldSessionId)
                fixed (byte* msgPtr = messageBuf)
                {
                    int msgLen = Event_WorldLeave_Query(_handle, wsidPtr, &code, msgPtr, messageBuf.Length);
                    if (msgLen < 0) return null;
                    return new WorldLeaveEventData
                    {
                        WorldSessionId = worldSessionId,
                        Code = code,
                        Message = Encoding.UTF8.GetString(messageBuf, 0, msgLen)
                    };
                }
            }
        }

        public PeerConnectedEventData? QueryPeerConnected()
        {
            if (Type != EventType.PeerConnected) return null;
            unsafe
            {
                byte[] peerIdBuf = new byte[256];
                IntPtr peerHandle;
                fixed (byte* pidPtr = peerIdBuf)
                {
                    int peerIdLen = Event_PeerConnected_Query(_handle, &peerHandle, pidPtr, peerIdBuf.Length);
                    if (peerIdLen < 0) return null;
                    return new PeerConnectedEventData
                    {
                        Peer = new Peer(peerHandle),
                        PeerId = Encoding.UTF8.GetString(peerIdBuf, 0, peerIdLen)
                    };
                }
            }
        }

        public string? QueryPeerDisconnected()
        {
            if (Type != EventType.PeerDisconnected) return null;
            unsafe
            {
                byte[] peerIdBuf = new byte[256];
                fixed (byte* pidPtr = peerIdBuf)
                {
                    int peerIdLen = Event_PeerDisconnected_Query(_handle, pidPtr, peerIdBuf.Length);
                    if (peerIdLen < 0) return null;
                    return Encoding.UTF8.GetString(peerIdBuf, 0, peerIdLen);
                }
            }
        }

        public void Dispose()
        {
            if (_disposed) return;
            if (_handle != IntPtr.Zero)
            {
                CloseEvent(_handle);
                _handle = IntPtr.Zero;
            }
            _disposed = true;
            GC.SuppressFinalize(this);
        }

        ~Event() => Dispose();
    }

    #region Event Data Structures

    public class WorldEnterEventData
    {
        public byte[] WorldSessionId { get; init; } = [];
        public string Url { get; init; } = string.Empty;
    }

    public class SessionEventData
    {
        public byte[] WorldSessionId { get; init; } = [];
        public byte[] PeerSessionId { get; init; } = [];
        public string PeerId { get; init; } = string.Empty;
    }

    public class ObjectInfo
    {
        public byte[] Id { get; init; } = [];
        public float[] Transform { get; init; } = [];
        public string Address { get; init; } = string.Empty;
    }

    public class ObjectAppendEventData : SessionEventData
    {
        public ObjectInfo[] Objects { get; set; } = [];
    }

    public class ObjectDeleteEventData : SessionEventData
    {
        public byte[][] ObjectIds { get; set; } = [];
    }

    public class WorldLeaveEventData
    {
        public byte[] WorldSessionId { get; init; } = [];
        public int Code { get; init; }
        public string Message { get; init; } = string.Empty;
    }

    public class PeerConnectedEventData
    {
        public Peer Peer { get; init; } = null!;
        public string PeerId { get; init; } = string.Empty;
    }

    #endregion

    #endregion

    #region Peer

    public class Peer : IDisposable
    {
        private IntPtr _handle;
        private bool _disposed;

        internal Peer(IntPtr handle) => _handle = handle;

        public bool IsValid => _handle != IntPtr.Zero;
        internal IntPtr Handle => _handle;

        public void Dispose()
        {
            if (_disposed) return;
            if (_handle != IntPtr.Zero)
            {
                ClosePeer(_handle);
                _handle = IntPtr.Zero;
            }
            _disposed = true;
            GC.SuppressFinalize(this);
        }

        ~Peer() => Dispose();
    }

    #endregion

    #region World

    public class World : IDisposable
    {
        private IntPtr _handle;
        private readonly Host _host;
        private bool _disposed;

        internal World(IntPtr handle, Host host)
        {
            _handle = handle;
            _host = host;
        }

        public bool IsValid => _handle != IntPtr.Zero;
        internal IntPtr Handle => _handle;

        public void AcceptSession(Peer peer, byte[] peerSessionId)
        {
            unsafe
            {
                fixed (byte* psidPtr = peerSessionId)
                {
                    World_AcceptSession(_host.Handle, _handle, peer.Handle, psidPtr);
                }
            }
        }

        public void DeclineSession(Peer peer, byte[] peerSessionId, int code, string message)
        {
            byte[] msgBytes = Encoding.UTF8.GetBytes(message);
            unsafe
            {
                fixed (byte* psidPtr = peerSessionId)
                fixed (byte* msgPtr = msgBytes)
                {
                    World_DeclineSession(_host.Handle, _handle, peer.Handle, psidPtr, code, msgPtr, msgBytes.Length);
                }
            }
        }

        public void Close() => World_Close(_host.Handle, _handle);

        public void Dispose()
        {
            if (_disposed) return;
            if (_handle != IntPtr.Zero)
            {
                CloseWorld(_handle);
                _handle = IntPtr.Zero;
            }
            _disposed = true;
            GC.SuppressFinalize(this);
        }

        ~World() => Dispose();
    }

    #endregion

    #region AbystClient

    public class AbystClient : IDisposable
    {
        private IntPtr _handle;
        private bool _disposed;

        internal AbystClient(IntPtr handle) => _handle = handle;

        public bool IsValid => _handle != IntPtr.Zero;

        public (HttpIOResult?, Error?) Get(IntPtr hEvent, string peerId, string path)
        {
            byte[] peerIdBytes = Encoding.UTF8.GetBytes(peerId);
            byte[] pathBytes = Encoding.UTF8.GetBytes(path);
            unsafe
            {
                fixed (byte* peerIdPtr = peerIdBytes)
                fixed (byte* pathPtr = pathBytes)
                {
                    IntPtr resultHandle;
                    IntPtr errHandle = AbystClient_Get(_handle, hEvent, peerIdPtr, peerIdBytes.Length, pathPtr, pathBytes.Length, &resultHandle);
                    if (errHandle != IntPtr.Zero)
                        return (null, new Error(errHandle));
                    return (new HttpIOResult(resultHandle), null);
                }
            }
        }

        public (HttpIOResult?, Error?) Post(IntPtr hEvent, string peerId, string path, string contentType, byte[] body)
        {
            byte[] peerIdBytes = Encoding.UTF8.GetBytes(peerId);
            byte[] pathBytes = Encoding.UTF8.GetBytes(path);
            byte[] contentTypeBytes = Encoding.UTF8.GetBytes(contentType);
            unsafe
            {
                fixed (byte* peerIdPtr = peerIdBytes)
                fixed (byte* pathPtr = pathBytes)
                fixed (byte* ctPtr = contentTypeBytes)
                fixed (byte* bodyPtr = body)
                {
                    IntPtr resultHandle;
                    IntPtr errHandle = AbystClient_Post(_handle, hEvent, peerIdPtr, peerIdBytes.Length, pathPtr, pathBytes.Length, ctPtr, contentTypeBytes.Length, bodyPtr, body.Length, &resultHandle);
                    if (errHandle != IntPtr.Zero)
                        return (null, new Error(errHandle));
                    return (new HttpIOResult(resultHandle), null);
                }
            }
        }

        public (HttpIOResult?, Error?) Head(IntPtr hEvent, string peerId, string path)
        {
            byte[] peerIdBytes = Encoding.UTF8.GetBytes(peerId);
            byte[] pathBytes = Encoding.UTF8.GetBytes(path);
            unsafe
            {
                fixed (byte* peerIdPtr = peerIdBytes)
                fixed (byte* pathPtr = pathBytes)
                {
                    IntPtr resultHandle;
                    IntPtr errHandle = AbystClient_Head(_handle, hEvent, peerIdPtr, peerIdBytes.Length, pathPtr, pathBytes.Length, &resultHandle);
                    if (errHandle != IntPtr.Zero)
                        return (null, new Error(errHandle));
                    return (new HttpIOResult(resultHandle), null);
                }
            }
        }

        public void Dispose()
        {
            if (_disposed) return;
            if (_handle != IntPtr.Zero)
            {
                CloseAbyssClient(_handle);
                _handle = IntPtr.Zero;
            }
            _disposed = true;
            GC.SuppressFinalize(this);
        }

        ~AbystClient() => Dispose();
    }

    #endregion

    #region Http3Client

    public class Http3Client : IDisposable
    {
        private IntPtr _handle;
        private bool _disposed;

        internal Http3Client(IntPtr handle) => _handle = handle;

        public bool IsValid => _handle != IntPtr.Zero;

        public (HttpIOResult?, Error?) Get(IntPtr hEvent, string url)
        {
            byte[] urlBytes = Encoding.UTF8.GetBytes(url);
            unsafe
            {
                fixed (byte* urlPtr = urlBytes)
                {
                    IntPtr resultHandle;
                    IntPtr errHandle = Http3Client_Get(_handle, hEvent, urlPtr, urlBytes.Length, &resultHandle);
                    if (errHandle != IntPtr.Zero)
                        return (null, new Error(errHandle));
                    return (new HttpIOResult(resultHandle), null);
                }
            }
        }

        public (HttpIOResult?, Error?) Post(IntPtr hEvent, string url, string contentType, byte[] body)
        {
            byte[] urlBytes = Encoding.UTF8.GetBytes(url);
            byte[] contentTypeBytes = Encoding.UTF8.GetBytes(contentType);
            unsafe
            {
                fixed (byte* urlPtr = urlBytes)
                fixed (byte* ctPtr = contentTypeBytes)
                fixed (byte* bodyPtr = body)
                {
                    IntPtr resultHandle;
                    IntPtr errHandle = Http3Client_Post(_handle, hEvent, urlPtr, urlBytes.Length, ctPtr, contentTypeBytes.Length, bodyPtr, body.Length, &resultHandle);
                    if (errHandle != IntPtr.Zero)
                        return (null, new Error(errHandle));
                    return (new HttpIOResult(resultHandle), null);
                }
            }
        }

        public (HttpIOResult?, Error?) Head(IntPtr hEvent, string url)
        {
            byte[] urlBytes = Encoding.UTF8.GetBytes(url);
            unsafe
            {
                fixed (byte* urlPtr = urlBytes)
                {
                    IntPtr resultHandle;
                    IntPtr errHandle = Http3Client_Head(_handle, hEvent, urlPtr, urlBytes.Length, &resultHandle);
                    if (errHandle != IntPtr.Zero)
                        return (null, new Error(errHandle));
                    return (new HttpIOResult(resultHandle), null);
                }
            }
        }

        public void Dispose()
        {
            if (_disposed) return;
            if (_handle != IntPtr.Zero)
            {
                CloseAbyssClientCollocatedHttp3Client(_handle);
                _handle = IntPtr.Zero;
            }
            _disposed = true;
            GC.SuppressFinalize(this);
        }

        ~Http3Client() => Dispose();
    }

    #endregion

    #region HttpIOResult

    public class HttpIOResult : IDisposable
    {
        private IntPtr _handle;
        private bool _disposed;

        internal HttpIOResult(IntPtr handle) => _handle = handle;

        public bool IsValid => _handle != IntPtr.Zero;

        public (HttpResponse?, Error?) Unpack()
        {
            unsafe
            {
                IntPtr responseHandle;
                IntPtr errHandle = HttpIOResult_Unpack(_handle, &responseHandle);
                if (errHandle != IntPtr.Zero)
                    return (null, new Error(errHandle));
                return (new HttpResponse(responseHandle), null);
            }
        }

        public void Dispose()
        {
            if (_disposed) return;
            if (_handle != IntPtr.Zero)
            {
                CloseHttpIOResult(_handle);
                _handle = IntPtr.Zero;
            }
            _disposed = true;
            GC.SuppressFinalize(this);
        }

        ~HttpIOResult() => Dispose();
    }

    public class HttpResponse : IDisposable
    {
        private IntPtr _handle;
        private bool _disposed;

        internal HttpResponse(IntPtr handle) => _handle = handle;

        public bool IsValid => _handle != IntPtr.Zero;

        public int StatusCode => HttpResponse_StatusCode(_handle);

        public string? GetHeader(string key)
        {
            byte[] keyBytes = Encoding.UTF8.GetBytes(key);
            byte[] valueBuf = new byte[1024];
            unsafe
            {
                fixed (byte* keyPtr = keyBytes)
                fixed (byte* valuePtr = valueBuf)
                {
                    int len = HttpResponse_GetHeader(_handle, keyPtr, keyBytes.Length, valuePtr, valueBuf.Length);
                    if (len < 0) return null;
                    return Encoding.UTF8.GetString(valueBuf, 0, len);
                }
            }
        }

        public string GetAllHeaders()
        {
            byte[] buf = new byte[8192];
            unsafe
            {
                fixed (byte* bufPtr = buf)
                {
                    int len = HttpResponse_GetAllHeaders(_handle, bufPtr, buf.Length);
                    if (len < 0) return string.Empty;
                    return Encoding.UTF8.GetString(buf, 0, len);
                }
            }
        }

        public int ReadBody(byte[] buffer)
        {
            unsafe
            {
                fixed (byte* bufPtr = buffer)
                {
                    return HttpResponse_ReadBody(_handle, bufPtr, buffer.Length);
                }
            }
        }

        public byte[] ReadAllBody()
        {
            List<byte> allBytes = [];
            byte[] buffer = new byte[8192];
            int bytesRead;
            while ((bytesRead = ReadBody(buffer)) > 0)
            {
                for (int i = 0; i < bytesRead; i++)
                    allBytes.Add(buffer[i]);
            }
            return [.. allBytes];
        }

        public void Dispose()
        {
            if (_disposed) return;
            if (_handle != IntPtr.Zero)
            {
                CloseHttpResponse(_handle);
                _handle = IntPtr.Zero;
            }
            _disposed = true;
            GC.SuppressFinalize(this);
        }

        ~HttpResponse() => Dispose();
    }

    #endregion
}
