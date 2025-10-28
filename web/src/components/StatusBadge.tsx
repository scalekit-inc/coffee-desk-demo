import { Badge } from "@/components/ui/badge";
import { cn } from "@/lib/utils";

interface StatusBadgeProps {
  status: 'Backlog' | 'Todo' | 'InProgress' | 'Done';
  className?: string;
}

export function StatusBadge({ status, className }: StatusBadgeProps) {
  const statusConfig = {
    Backlog: {
      label: 'Backlog',
      variant: 'secondary' as const,
      className: 'bg-gray-100 text-gray-800 hover:bg-gray-200',
    },
    Todo: {
      label: 'To Do',
      variant: 'outline' as const,
      className: 'bg-blue-50 text-blue-700 border-blue-200 hover:bg-blue-100',
    },
    InProgress: {
      label: 'In Progress',
      variant: 'default' as const,
      className: 'bg-yellow-100 text-yellow-800 hover:bg-yellow-200',
    },
    Done: {
      label: 'Done',
      variant: 'default' as const,
      className: 'bg-green-100 text-green-800 hover:bg-green-200',
    },
  };

  const config = statusConfig[status];

  return (
    <Badge
      variant={config.variant}
      className={cn(config.className, className)}
    >
      {config.label}
    </Badge>
  );
}
