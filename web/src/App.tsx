import { useTasks } from '@/hooks/useTasks';
import { useSSE } from '@/hooks/useSSE';
import { Header } from '@/components/Header';
import { KanbanBoard } from '@/components/KanbanBoard';

function App() {
  const {
    data,
    loading,
    error,
    currentProject,
    refresh,
    addTask,
    moveTask,
    removeTask,
    editTask,
    switchProject,
  } = useTasks();

  // Listen for SSE updates
  useSSE(refresh);

  if (loading && !data) {
    return (
      <div className="min-h-screen bg-background flex items-center justify-center">
        <div className="text-muted-foreground">Loading...</div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="min-h-screen bg-background flex items-center justify-center">
        <div className="text-red-500">Error: {error}</div>
      </div>
    );
  }

  if (!data) {
    return null;
  }

  return (
    <div className="min-h-screen bg-background">
      <div className="container mx-auto px-4 py-6 max-w-7xl">
        <Header
          projects={data.projects}
          currentProject={currentProject}
          onProjectChange={switchProject}
        />
        <main>
          <KanbanBoard
            data={data}
            onAddTask={addTask}
            onMoveTask={moveTask}
            onDeleteTask={removeTask}
            onEditTask={editTask}
          />
        </main>
        <footer className="mt-8 pt-4 border-t text-center text-sm text-muted-foreground">
          mini-kanban — A simple kanban board 2026
        </footer>
      </div>
    </div>
  );
}

export default App;
