using AbyssCLI.ABI;
using System;
using System.Reflection;
using System.Text;

namespace Host
{
    partial class Host
    {
        private void LogRequest(RenderAction render_action)
        {
            switch (render_action.InnerCase)
            {
            case RenderAction.InnerOneofCase.ConsolePrint: GlobalDependency.Logger.Writer.WriteLine(FormatFlatLogLine(render_action.ConsolePrint)); return;
            case RenderAction.InnerOneofCase.CreateElement: GlobalDependency.Logger.Writer.WriteLine(FormatFlatLogLine(render_action.CreateElement)); return;
            case RenderAction.InnerOneofCase.MoveElement: GlobalDependency.Logger.Writer.WriteLine(FormatFlatLogLine(render_action.MoveElement)); return;
            case RenderAction.InnerOneofCase.DeleteElement: GlobalDependency.Logger.Writer.WriteLine(FormatFlatLogLine(render_action.DeleteElement)); return;
            case RenderAction.InnerOneofCase.ElemSetActive: GlobalDependency.Logger.Writer.WriteLine(FormatFlatLogLine(render_action.ElemSetActive)); return;
            case RenderAction.InnerOneofCase.ElemSetTransform: GlobalDependency.Logger.Writer.WriteLine(FormatFlatLogLine(render_action.ElemSetTransform)); return;
            case RenderAction.InnerOneofCase.ElemAttachResource: GlobalDependency.Logger.Writer.WriteLine(FormatFlatLogLine(render_action.ElemAttachResource)); return;
            case RenderAction.InnerOneofCase.ElemDetachResource: GlobalDependency.Logger.Writer.WriteLine(FormatFlatLogLine(render_action.ElemDetachResource)); return;
            case RenderAction.InnerOneofCase.CreateItem: GlobalDependency.Logger.Writer.WriteLine(FormatFlatLogLine(render_action.CreateItem)); return;
            case RenderAction.InnerOneofCase.DeleteItem: GlobalDependency.Logger.Writer.WriteLine(FormatFlatLogLine(render_action.DeleteItem)); return;
            case RenderAction.InnerOneofCase.ItemSetTitle: GlobalDependency.Logger.Writer.WriteLine(FormatFlatLogLine(render_action.ItemSetTitle)); return;
            case RenderAction.InnerOneofCase.ItemSetIcon: GlobalDependency.Logger.Writer.WriteLine(FormatFlatLogLine(render_action.ItemSetIcon)); return;
            case RenderAction.InnerOneofCase.ItemSetActive: GlobalDependency.Logger.Writer.WriteLine(FormatFlatLogLine(render_action.ItemSetActive)); return;
            case RenderAction.InnerOneofCase.ItemAlert: GlobalDependency.Logger.Writer.WriteLine(FormatFlatLogLine(render_action.ItemAlert)); return;
            case RenderAction.InnerOneofCase.OpenStaticResource: GlobalDependency.Logger.Writer.WriteLine(FormatFlatLogLine(render_action.OpenStaticResource)); return;
            case RenderAction.InnerOneofCase.CloseResource: GlobalDependency.Logger.Writer.WriteLine(FormatFlatLogLine(render_action.CloseResource)); return;
            case RenderAction.InnerOneofCase.CreateCompositeResource: GlobalDependency.Logger.Writer.WriteLine(FormatFlatLogLine(render_action.CreateCompositeResource)); return;
            case RenderAction.InnerOneofCase.MemberInfo: GlobalDependency.Logger.Writer.WriteLine(FormatFlatLogLine(render_action.MemberInfo)); return;
            case RenderAction.InnerOneofCase.MemberSetProfile: GlobalDependency.Logger.Writer.WriteLine(FormatFlatLogLine(render_action.MemberSetProfile)); return;
            case RenderAction.InnerOneofCase.MemberLeave: GlobalDependency.Logger.Writer.WriteLine(FormatFlatLogLine(render_action.MemberLeave)); return;
            case RenderAction.InnerOneofCase.LocalInfo: GlobalDependency.Logger.Writer.WriteLine(FormatFlatLogLine(render_action.LocalInfo)); return;
            case RenderAction.InnerOneofCase.InfoContentShared: GlobalDependency.Logger.Writer.WriteLine(FormatFlatLogLine(render_action.InfoContentShared)); return;
            case RenderAction.InnerOneofCase.InfoContentDeleted: GlobalDependency.Logger.Writer.WriteLine(FormatFlatLogLine(render_action.InfoContentDeleted)); return;
            case RenderAction.InnerOneofCase.DebugEnter: return;
            case RenderAction.InnerOneofCase.DebugLeave: return;
            default: StderrQueue.Enqueue("Executor: invalid RenderAction: " + render_action.InnerCase); return;
            }
        }
        string FormatFlatLogLine(object obj)
        {
            var sb = new StringBuilder();
            //_ = sb.Append($"[{DateTime.Now:yyyy-MM-dd HH:mm:ss}] {obj.GetType().Name} |");
            _ = sb.Append($"{obj.GetType().Name} |");

            var type = obj.GetType();
            var fields = type.GetFields(BindingFlags.Instance | BindingFlags.Public);
            var properties = type.GetProperties(BindingFlags.Instance | BindingFlags.Public);

            foreach (var field in fields)
            {
                if (!IsSimple(field.FieldType)) continue;
                _ = sb.Append($" {field.Name}={FormatValue(field.GetValue(obj))}");
            }

            foreach (var prop in properties)
            {
                if (!prop.CanRead || !IsSimple(prop.PropertyType)) continue;
                _ = sb.Append($" {prop.Name}={FormatValue(prop.GetValue(obj))}");
            }

            return sb.ToString();
        }
        bool IsSimple(Type type)
        {
            return type.IsPrimitive || type == typeof(string) || type == typeof(byte[]);
        }

        string FormatValue(object value)
        {
            if (value == null) return "null";

            return value switch
            {
                string s => s,
                byte[] bytes => BitConverter.ToString(bytes).Replace("-", ""), // Hex string
                bool b => b ? "true" : "false",
                float f => f.ToString("R"),
                double d => d.ToString("R"),
                _ => value.ToString()
            };
        }
    }
}