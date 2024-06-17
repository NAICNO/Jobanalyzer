interface JobQueryResultsTableColumnHeader {
  key: string;
  title: string;
  shortTitle?: string;
  helpText?: string;
  sortable?: boolean;
  description?: string;
  formatterFns?: ((value: any) => any)[];
  renderFn?: (value: any) => any;
  minSize?: number;
  sortingFn?: (rowA: any, rowB: any, columnId: any) => number;
}
