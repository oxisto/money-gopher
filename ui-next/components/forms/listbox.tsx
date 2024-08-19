interface ListBoxProps {
  value: any;
  name: string;
  options: { value: any; display: string }[];
}

export default function ListBox({ name, value, options }: ListBoxProps) {
  return (
    <>
      <input type="hidden" name={name} defaultValue={value} />
      {value}
    </>
  );
}
