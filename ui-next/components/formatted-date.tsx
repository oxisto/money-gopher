interface DateProps {
  /**
   * The date to format.
   */
  date?: Date;
}

export default function FormattedDate({ date }: DateProps) {
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
