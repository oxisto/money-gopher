import { ChangeEvent, useState } from "react";
import dayjs from "dayjs";
import { dateTimeLocalFormat } from "@/lib/util";

interface DateInputProps {
  value: Date | string;
  name: string;
  onChange?: (
    /**
     * The date in ISO-8601.
     */
    value: string
  ) => void;
}

export default function DateInput({ name, value, onChange }: DateInputProps) {
  function onDateChanged(e: ChangeEvent<HTMLInputElement>) {
    // update our internal value...
    setInternal(e.target.value);

    const isoDate = dayjs(e.target.value, dateTimeLocalFormat).toISOString();
    onChange?.call(null, isoDate);
  }

  if (typeof value === "string") {
    value = new Date(Date.parse(value));
  }

  let [internal, setInternal] = useState(
    dayjs(value).format(dateTimeLocalFormat)
  );

  return (
    <input
      type="datetime-local"
      name={name}
      id={name}
      className="block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300
  placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6"
      value={internal}
      onChange={onDateChanged}
    />
  );
}
