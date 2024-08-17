interface DebugProps {
  /**
   * The object to display.
   */
  obj: any;
}
export default function Debug({ obj }: DebugProps) {
  return (
    <div className="font-mono">
      <pre>{JSON.stringify(obj, null, 2)}</pre>
    </div>
  );
}
