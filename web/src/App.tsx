import { useState, useCallback, useEffect, useRef } from 'react';
import { useTasks } from '@/hooks/useTasks';
import { useSSE, type SSENotification } from '@/hooks/useSSE';
import { useNotification } from '@/hooks/useNotification';
import { useTranslation } from '@/hooks/useTranslation';
import { Header } from '@/components/Header';
import { KanbanBoard } from '@/components/KanbanBoard';
import { ConfigDialog } from '@/components/ConfigDialog';

function App() {
  const { t } = useTranslation();
  const [configOpen, setConfigOpen] = useState(false);
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

  const { permission, requestPermission, showNotification, isSupported } = useNotification();
  const notifiedOverdueRef = useRef<Set<number>>(new Set());

  const handleNotification = useCallback((notification: SSENotification) => {
    if (notification.type === 'status_change') {
      if (notification.newStatus === 'review') {
        showNotification(t('notificationStatusReview'), {
          body: t('notificationBody', { title: notification.taskTitle }),
          icon: '/favicon.ico',
        });
      } else if (notification.newStatus === 'done') {
        showNotification(t('notificationStatusDone'), {
          body: t('notificationBody', { title: notification.taskTitle }),
          icon: '/favicon.ico',
        });
      }
    } else if (notification.type === 'overdue') {
      showNotification(t('notificationOverdue'), {
        body: t('notificationBody', { title: notification.taskTitle }),
        icon: '/favicon.ico',
      });
    }
  }, [showNotification, t]);

  useSSE(refresh, handleNotification);

  useEffect(() => {
    if (!data) return;

    const checkOverdue = () => {
      const now = Date.now() / 1000;
      const allTasks = [...data.todoTasks, ...data.doingTasks, ...data.reviewTasks];
      
      for (const task of allTasks) {
        if (task.dueAt && task.dueAt < now && !notifiedOverdueRef.current.has(task.id)) {
          notifiedOverdueRef.current.add(task.id);
          showNotification(t('notificationOverdue'), {
            body: t('notificationBody', { title: task.title }),
            icon: '/favicon.ico',
          });
        }
      }
    };

    checkOverdue();
    const interval = setInterval(checkOverdue, 60000);

    return () => clearInterval(interval);
  }, [data, showNotification, t]);

  useEffect(() => {
    if (isSupported && permission === 'default') {
      requestPermission();
    }
  }, [isSupported, permission, requestPermission]);

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
          onSettingsClick={() => setConfigOpen(true)}
          onProjectCreated={refresh}
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
      <ConfigDialog open={configOpen} onOpenChange={setConfigOpen} />
    </div>
  );
}

export default App;
