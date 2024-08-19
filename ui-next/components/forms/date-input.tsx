interface DateInputProps {
  value: Date | string;
  name: string;
}

export default function DateInput({ name, value }: DateInputProps) {
  if (typeof value === "string") {
    value = new Date(Date.parse(value));
  }

  return (
    <>
      <input type="hidden" name={name} defaultValue={value.toISOString()} />
      {value.toISOString()}
    </>
  );
}
