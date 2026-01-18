import { useState } from 'react';
import { useTranslation } from '@/hooks/useTranslation';
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
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';

interface TaskCardProps {
  task: Task;
  showProject?: boolean;
  availableTags?: string[];
  onMove: (taskId: number, newStatus: Task['status'], project: string) => void;
  onDelete: (taskId: number, project: string) => void;
  onEdit: (taskId: number, updates: { title?: string; body?: string; tags?: string[]; dueAt?: string; project?: string }) => void;
}

export function TaskCard({ task, showProject, availableTags, onMove, onDelete, onEdit }: TaskCardProps) {
  const { t } = useTranslation();
  const [editOpen, setEditOpen] = useState(false);
  const [deleteConfirmOpen, setDeleteConfirmOpen] = useState(false);

  const formatDate = (unix: number) => {
    const date = new Date(unix * 1000);
    return `${(date.getMonth() + 1).toString().padStart(2, '0')}/${date.getDate().toString().padStart(2, '0')}`;
  };

  const handleDragStart = (e: React.DragEvent) => {
    e.dataTransfer.setData('taskId', task.id.toString());
    e.dataTransfer.setData('project', task.project);
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
                {showProject && (
                  <Badge variant="outline" className="text-xs px-1 py-0">
                    {task.project}
                  </Badge>
                )}
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
                  <DropdownMenuItem onClick={() => onMove(task.id, 'todo', task.project)}>
                    {t('moveToTodo')}
                  </DropdownMenuItem>
                )}
                {task.status !== 'doing' && (
                  <DropdownMenuItem onClick={() => onMove(task.id, 'doing', task.project)}>
                    {t('moveToDoing')}
                  </DropdownMenuItem>
                )}
                {task.status !== 'review' && (
                  <DropdownMenuItem onClick={() => onMove(task.id, 'review', task.project)}>
                    {t('moveToReview')}
                  </DropdownMenuItem>
                )}
                {task.status !== 'done' && (
                  <DropdownMenuItem onClick={() => onMove(task.id, 'done', task.project)}>
                    {t('moveToDone')}
                  </DropdownMenuItem>
                )}
                <DropdownMenuItem onClick={() => setEditOpen(true)}>
                  {t('edit')}
                </DropdownMenuItem>
                <DropdownMenuItem 
                  className="text-red-600"
                  onClick={() => setDeleteConfirmOpen(true)}
                >
                  {t('delete')}
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
        project={task.project}
        availableTags={availableTags}
        open={editOpen}
        onOpenChange={setEditOpen}
        onSave={onEdit}
      />
      
      <Dialog open={deleteConfirmOpen} onOpenChange={setDeleteConfirmOpen}>
        <DialogContent showCloseButton={false}>
          <DialogHeader>
            <DialogTitle>{t('deleteConfirmTitle')}</DialogTitle>
            <DialogDescription>{t('deleteConfirm')}</DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button variant="outline" onClick={() => setDeleteConfirmOpen(false)}>
              {t('cancel')}
            </Button>
            <Button
              variant="destructive"
              onClick={() => {
                onDelete(task.id, task.project);
                setDeleteConfirmOpen(false);
              }}
            >
              {t('confirmDelete')}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  );
}
