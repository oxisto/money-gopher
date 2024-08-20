import { classNames } from "@/lib/util";

interface ButtonProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  children: React.ReactNode;
  className?: string;
}

export default function Button({
  type = "button",
  children,
  className,
  ...rest
}: ButtonProps) {
  return (
    <button
      type={type}
      className={classNames(
        "mt-5 rounded-md bg-indigo-600 px-3 py-2 text-sm font-semibold text-white shadow-sm hover:bg-indigo-500 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600",
        className ?? ''
      )}
      {...rest}
    >
      {children}
    </button>
  );
}
