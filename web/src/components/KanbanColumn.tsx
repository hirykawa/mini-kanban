import { useTranslation } from '@/hooks/useTranslation';
import type { Task } from '@/lib/api';
import { TaskCard } from './TaskCard';

interface KanbanColumnProps {
  title: string;
  status: Task['status'];
  tasks: Task[];
  showProject: boolean;
  availableTags?: string[];
  onMove: (taskId: number, newStatus: Task['status'], project: string) => void;
  onDelete: (taskId: number, project: string) => void;
  onEdit: (taskId: number, updates: { title?: string; body?: string; tags?: string[]; dueAt?: string; project?: string }) => void;
}

export function KanbanColumn({ title, status, tasks, showProject, availableTags, onMove, onDelete, onEdit }: KanbanColumnProps) {
  const { t } = useTranslation();
  const handleDragOver = (e: React.DragEvent) => {
    e.preventDefault();
    e.currentTarget.classList.add('bg-accent/50');
  };

  const handleDragLeave = (e: React.DragEvent) => {
    e.currentTarget.classList.remove('bg-accent/50');
  };

  const handleDrop = (e: React.DragEvent) => {
    e.preventDefault();
    e.currentTarget.classList.remove('bg-accent/50');
    const taskId = parseInt(e.dataTransfer.getData('taskId'), 10);
    const project = e.dataTransfer.getData('project');
    if (taskId && project) {
      onMove(taskId, status, project);
    }
  };

  const getColumnStyle = () => {
    switch (status) {
      case 'todo':
        return 'border-t-blue-500';
      case 'doing':
        return 'border-t-yellow-500';
      case 'review':
        return 'border-t-purple-500';
      case 'done':
        return 'border-t-green-500';
    }
  };

  return (
    <div className={`flex flex-col bg-muted/30 rounded-lg border-t-4 ${getColumnStyle()} min-h-[400px]`}>
      <div className="flex items-center justify-between p-4 pb-2">
        <h2 className="font-semibold text-lg">{title}</h2>
        <span className="px-2 py-1 text-sm bg-accent rounded-full">{tasks.length}</span>
      </div>
      <div
        className="flex-1 p-2 space-y-2 transition-colors"
        onDragOver={handleDragOver}
        onDragLeave={handleDragLeave}
        onDrop={handleDrop}
      >
        {tasks.map((task) => (
          <TaskCard
            key={`${task.project}-${task.id}`}
            task={task}
            showProject={showProject}
            availableTags={availableTags}
            onMove={onMove}
            onDelete={onDelete}
            onEdit={onEdit}
          />
        ))}
        {tasks.length === 0 && (
          <div className="text-center text-muted-foreground text-sm py-8">
            {t('noTasks')}
          </div>
        )}
      </div>
    </div>
  );
}
