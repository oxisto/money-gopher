import {
  Listbox,
  ListboxButton,
  ListboxOption,
  ListboxOptions,
  ListboxProps,
} from "@headlessui/react";
import { CheckIcon, ChevronUpDownIcon } from "@heroicons/react/24/outline";
import { useState } from "react";

interface ListBoxProps<T>
  extends ListboxProps<
    React.ExoticComponent<{ children?: React.ReactNode }>,
    T
  > {
  value: T;
  name: string;
  options: { value: T; display: string }[];
}

export default function ListBox<T>({
  name,
  value,
  options,
  onChange,
}: ListBoxProps<T>) {
  const [selected, setSelected] = useState<
    | {
      value: T;
      display: string;
    }
    | undefined
  >(options.find((option) => option.value == value));
  return (
    <Listbox
      name={name}
      value={selected}
      onChange={(value) => {
        setSelected(value);
        onChange?.call(null, value.value);
      }}
    >
      <div className="relative mt-2">
        <ListboxButton
          className="relative w-full cursor-default rounded-md bg-white py-1.5 pl-3 pr-10 text-left
            text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 focus:outline-none
            focus:ring-2 focus:ring-indigo-500 sm:text-sm sm:leading-6"
        >
          <span className="inline-flex w-full truncate">
            {selected != null ? (
              <span className="truncate">{selected.display}</span>
            ) : (
              <span className="truncate">Please select</span>
            )}
          </span>
          <span className="pointer-events-none absolute inset-y-0 right-0 flex items-center pr-2">
            <ChevronUpDownIcon
              className="h-5 w-5 text-gray-400"
              aria-hidden="true"
            />
          </span>
        </ListboxButton>
        <ListboxOptions
          transition
          className="absolute z-10 mt-1 max-h-56 w-full overflow-auto rounded-md bg-white py-1
            text-base shadow-lg ring-1 ring-black ring-opacity-5 focus:outline-none
            data-[closed]:data-[leave]:opacity-0 data-[leave]:transition
            data-[leave]:duration-100 data-[leave]:ease-in sm:text-sm"
        >
          {options.map((option, idx) => (
            <ListboxOption
              key={idx}
              value={option}
              className="group relative cursor-default select-none py-2 pl-3 pr-9 text-gray-900
                data-[focus]:bg-indigo-600 data-[focus]:text-white"
            >
              <div className="flex items-center">
                <span className="ml-3 block truncate font-normal group-data-[selected]:font-semibold">
                  {option.display}
                </span>
              </div>

              <span
                className="absolute inset-y-0 right-0 flex items-center pr-4 text-indigo-600
                  group-data-[focus]:text-white [.group:not([data-selected])_&]:hidden"
              >
                <CheckIcon aria-hidden="true" className="h-5 w-5" />
              </span>
            </ListboxOption>
          ))}
        </ListboxOptions>
      </div>
    </Listbox>
  );
}
