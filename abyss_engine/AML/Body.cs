namespace AbyssCLI.AML;

public class Body : Transform
{
    public Body(Document document) : base(document, "body", null) {}
    public void Init()
    {
        Client.Client.RenderWriter.MoveElement(ElementId, 0);
    }

    public override void remove() =>
        throw new InvalidOperationException("<body> cannot be removed");
}
