export interface DashboardTableColumnHeader {
  key: string;
  title: string;
  shortTitle?: string;
  helpText?: string;
  sortable?: boolean;
  description?: string;
  renderFn?: (value: any) => any;
}
