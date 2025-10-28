import { Badge } from "@/components/ui/badge";
import { cn } from "@/lib/utils";

interface PriorityBadgeProps {
  priority: 'P1' | 'P2' | 'P3';
  className?: string;
}

export function PriorityBadge({ priority, className }: PriorityBadgeProps) {
  const priorityConfig = {
    P1: {
      label: 'P1',
      className: 'bg-red-100 text-red-800 hover:bg-red-200',
    },
    P2: {
      label: 'P2',
      className: 'bg-orange-100 text-orange-800 hover:bg-orange-200',
    },
    P3: {
      label: 'P3',
      className: 'bg-green-100 text-green-800 hover:bg-green-200',
    },
  };

  const config = priorityConfig[priority];

  return (
    <Badge
      variant="outline"
      className={cn(config.className, className)}
    >
      {config.label}
    </Badge>
  );
}
