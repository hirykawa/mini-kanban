import { useTranslation } from '@/hooks/useTranslation';
import type { KanbanData, Task } from '@/lib/api';
import { KanbanColumn } from './KanbanColumn';
import { TaskForm } from './TaskForm';

interface KanbanBoardProps {
  data: KanbanData;
  isMultiProject: boolean;
  onAddTask: (title: string, tags?: string[], dueAt?: string, project?: string) => void;
  onMoveTask: (taskId: number, newStatus: Task['status'], project: string) => void;
  onDeleteTask: (taskId: number, project: string) => void;
  onEditTask: (taskId: number, updates: { title?: string; body?: string; tags?: string[]; dueAt?: string; project?: string }) => void;
}

export function KanbanBoard({ data, isMultiProject, onAddTask, onMoveTask, onDeleteTask, onEditTask }: KanbanBoardProps) {
  const { t } = useTranslation();

  return (
    <div className="space-y-6">
      <TaskForm
        project={data.currentProject}
        projects={isMultiProject ? data.currentProjects : undefined}
        availableTags={data.tags}
        onSubmit={onAddTask}
      />
      
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
        <KanbanColumn
          title={t('todo')}
          status="todo"
          tasks={data.todoTasks || []}
          showProject={isMultiProject}
          availableTags={data.tags}
          onMove={onMoveTask}
          onDelete={onDeleteTask}
          onEdit={onEditTask}
        />
        <KanbanColumn
          title={t('doing')}
          status="doing"
          tasks={data.doingTasks || []}
          showProject={isMultiProject}
          availableTags={data.tags}
          onMove={onMoveTask}
          onDelete={onDeleteTask}
          onEdit={onEditTask}
        />
        <KanbanColumn
          title={t('preview')}
          status="review"
          tasks={data.reviewTasks || []}
          showProject={isMultiProject}
          availableTags={data.tags}
          onMove={onMoveTask}
          onDelete={onDeleteTask}
          onEdit={onEditTask}
        />
        <KanbanColumn
          title={t('done')}
          status="done"
          tasks={data.doneTasks || []}
          showProject={isMultiProject}
          availableTags={data.tags}
          onMove={onMoveTask}
          onDelete={onDeleteTask}
          onEdit={onEditTask}
        />
      </div>
    </div>
  );
}
