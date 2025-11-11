using System.Collections.Concurrent;

namespace AbyssCLI.Tool;

public class SingleThreadTaskRunner
{
    private readonly BlockingCollection<Func<Task>> _queue = [];
    private readonly Thread _thread;
    private readonly CancellationTokenSource _cts = new();

    //TODO: manipulate scheduling
    public SingleThreadTaskRunner()
    {
        _thread = new Thread(MainLoop);
    }
    private void MainLoop()
    {
        //SynchronizationContext.SetSynchronizationContext(SynchronizationContext.Current);
        try
        {
            foreach (Func<Task> work in _queue.GetConsumingEnumerable(_cts.Token))
            {
                _ = work();
            }
        }
        catch (OperationCanceledException) { }
    }
    public void Start() =>
        _thread.Start();
    public void Post(Func<Task> work) =>
        _queue.Add(work);
    public void Stop()
    {
        _cts.Cancel();
        _queue.CompleteAdding();
        _thread.Join();
        _queue.Dispose();
        _cts.Dispose();
    }
}
