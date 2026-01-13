import type { KanbanData, Task } from '@/lib/api';
import { KanbanColumn } from './KanbanColumn';
import { TaskForm } from './TaskForm';

interface KanbanBoardProps {
  data: KanbanData;
  onAddTask: (title: string, tags?: string[], dueAt?: string) => void;
  onMoveTask: (taskId: number, newStatus: Task['status']) => void;
  onDeleteTask: (taskId: number) => void;
  onEditTask: (taskId: number, updates: { title?: string; tags?: string[]; dueAt?: string }) => void;
}

export function KanbanBoard({ data, onAddTask, onMoveTask, onDeleteTask, onEditTask }: KanbanBoardProps) {
  return (
    <div className="space-y-6">
      <TaskForm project={data.currentProject} onSubmit={onAddTask} />
      
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        <KanbanColumn
          title="Todo"
          status="todo"
          tasks={data.todoTasks}
          project={data.currentProject}
          onMove={onMoveTask}
          onDelete={onDeleteTask}
          onEdit={onEditTask}
        />
        <KanbanColumn
          title="Doing"
          status="doing"
          tasks={data.doingTasks}
          project={data.currentProject}
          onMove={onMoveTask}
          onDelete={onDeleteTask}
          onEdit={onEditTask}
        />
        <KanbanColumn
          title="Done"
          status="done"
          tasks={data.doneTasks}
          project={data.currentProject}
          onMove={onMoveTask}
          onDelete={onDeleteTask}
          onEdit={onEditTask}
        />
      </div>
    </div>
  );
}
