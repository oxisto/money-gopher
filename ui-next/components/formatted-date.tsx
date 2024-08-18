interface DateProps {
  /**
   * The date to format.
   */
  date?: Date | string;
}

export default function FormattedDate({ date }: DateProps) {
  if (typeof date === "string") {
    date = new Date(Date.parse(date));
  }

  return (
    date && (
      <time dateTime={date.toISOString()}>
        {Intl.DateTimeFormat(navigator.language, {
          weekday: "long",
          year: "numeric",
          month: "long",
          day: "numeric",
        }).format(date)}
      </time>
    )
  );
}
