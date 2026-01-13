import { useState } from 'react';
import type { Task } from '@/lib/api';
import { Card, CardContent, CardHeader } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { TaskEditDialog } from './TaskEditDialog';

interface TaskCardProps {
  task: Task;
  project: string;
  onMove: (taskId: number, newStatus: Task['status']) => void;
  onDelete: (taskId: number) => void;
  onEdit: (taskId: number, updates: { title?: string; tags?: string[]; dueAt?: string }) => void;
}

export function TaskCard({ task, project, onMove, onDelete, onEdit }: TaskCardProps) {
  const [editOpen, setEditOpen] = useState(false);

  const formatDate = (unix: number) => {
    const date = new Date(unix * 1000);
    return `${(date.getMonth() + 1).toString().padStart(2, '0')}/${date.getDate().toString().padStart(2, '0')}`;
  };

  const handleDragStart = (e: React.DragEvent) => {
    e.dataTransfer.setData('taskId', task.id.toString());
    e.dataTransfer.effectAllowed = 'move';
  };

  return (
    <>
      <Card
        draggable
        onDragStart={handleDragStart}
        className={`cursor-grab active:cursor-grabbing transition-all hover:shadow-md ${
          task.isOverdue ? 'border-red-500 border-2' : ''
        } ${task.status === 'done' ? 'opacity-60' : ''}`}
      >
        <CardHeader className="p-3 pb-1">
          <div className="flex items-start justify-between gap-2">
            <div className="flex-1 min-w-0">
              <div className="flex items-center gap-2 mb-1">
                <span className="text-xs text-muted-foreground font-mono">#{task.id}</span>
                {task.dueAt && (
                  <span className={`text-xs ${task.isOverdue ? 'text-red-500 font-bold' : 'text-muted-foreground'}`}>
                    {formatDate(task.dueAt)}
                  </span>
                )}
              </div>
              <h3 
                className="font-medium text-sm leading-tight break-words cursor-pointer hover:text-blue-600"
                onClick={() => setEditOpen(true)}
              >
                {task.title}
              </h3>
            </div>
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <Button variant="ghost" size="sm" className="h-6 w-6 p-0 shrink-0">
                  ⋮
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end">
                {task.status !== 'todo' && (
                  <DropdownMenuItem onClick={() => onMove(task.id, 'todo')}>
                    Move to Todo
                  </DropdownMenuItem>
                )}
                {task.status !== 'doing' && (
                  <DropdownMenuItem onClick={() => onMove(task.id, 'doing')}>
                    Move to Doing
                  </DropdownMenuItem>
                )}
                {task.status !== 'done' && (
                  <DropdownMenuItem onClick={() => onMove(task.id, 'done')}>
                    Move to Done
                  </DropdownMenuItem>
                )}
                <DropdownMenuItem onClick={() => setEditOpen(true)}>
                  Edit
                </DropdownMenuItem>
                <DropdownMenuItem 
                  className="text-red-600"
                  onClick={() => {
                    if (confirm('Delete this task?')) {
                      onDelete(task.id);
                    }
                  }}
                >
                  Delete
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
          </div>
        </CardHeader>
        {task.tags.length > 0 && (
          <CardContent className="p-3 pt-1">
            <div className="flex flex-wrap gap-1">
              {task.tags.map((tag) => (
                <Badge key={tag} variant="secondary" className="text-xs">
                  {tag}
                </Badge>
              ))}
            </div>
          </CardContent>
        )}
      </Card>
      
      <TaskEditDialog
        task={task}
        project={project}
        open={editOpen}
        onOpenChange={setEditOpen}
        onSave={onEdit}
      />
    </>
  );
}
