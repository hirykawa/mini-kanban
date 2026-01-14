import { useTasks } from '@/hooks/useTasks';
import { useSSE } from '@/hooks/useSSE';
import { useTranslation } from '@/hooks/useTranslation';
import { Header } from '@/components/Header';
import { KanbanBoard } from '@/components/KanbanBoard';

function App() {
  const { t } = useTranslation();
  const {
    data,
    loading,
    error,
    currentProjects,
    isMultiProject,
    refresh,
    addTask,
    moveTask,
    removeTask,
    editTask,
    switchProjects,
  } = useTasks();

  // Listen for SSE updates
  useSSE(refresh);

  if (loading && !data) {
    return (
      <div className="min-h-screen bg-background flex items-center justify-center">
        <div className="text-muted-foreground">{t('loading')}</div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="min-h-screen bg-background flex items-center justify-center">
        <div className="text-red-500">{t('error', { error })}</div>
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
          currentProjects={currentProjects}
          onProjectChange={switchProjects}
        />
        <main>
          <KanbanBoard
            data={data}
            isMultiProject={isMultiProject}
            onAddTask={addTask}
            onMoveTask={moveTask}
            onDeleteTask={removeTask}
            onEditTask={editTask}
          />
        </main>
        <footer className="mt-8 pt-4 border-t text-center text-sm text-muted-foreground">
          {t('footer')}
        </footer>
      </div>
    </div>
  );
}

export default App;
