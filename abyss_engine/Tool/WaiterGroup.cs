namespace AbyssCLI.Tool;

[Obsolete("Use TaskCompletionSource")]
internal class WaiterGroup<T>
{
    public bool TryFinalizeValue(T value)
    {
        lock (_waiters)
        {
            if (finalized)
            {
                return false;
            }

            result = value;
            finalized = true;
            foreach (Waiter<T> waiter in _waiters)
            {
                waiter.Finalize(value);
            }
            _waiters.Clear();
            return true;
        }
    }
    public bool TryGetValueOrWaiter(out T value, out Waiter<T> waiter)
    {
        lock (_waiters)
        {
            if (finalized)
            {
                value = result;
                waiter = null;
                return true;
            }

            value = default;
            waiter = new Waiter<T>();
            _ = _waiters.Add(waiter);
            return false;
        }
    }
    public T GetValue() => result;
    public bool IsFinalized() => finalized;
    private T result;
    private bool finalized = false; //0: init, 1: loading, 2: loaded (no need to check sema)
    private readonly HashSet<Waiter<T>> _waiters = [];
}
