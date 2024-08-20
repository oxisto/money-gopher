interface TableSorterProps {
  active: boolean;
  column: string;
  children: React.ReactNode;
  onChangeDirection: (column: string) => (void);
  onChangeSortBy: (column: string) => (void);
}

export default function TableSorter({
  active,
  column,
  children,
}: TableSorterProps) {
  return <>{children}</>;
}
