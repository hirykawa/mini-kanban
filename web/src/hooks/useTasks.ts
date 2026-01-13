import { useState, useEffect, useCallback } from 'react';
import type { KanbanData, Task } from '@/lib/api';
import { fetchKanban, createTask, updateTaskStatus, deleteTask, updateTask } from '@/lib/api';

export function useTasks(initialProject?: string) {
  const [data, setData] = useState<KanbanData | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [currentProject, setCurrentProject] = useState(initialProject || '');

  const refresh = useCallback(async () => {
    try {
      setLoading(true);
      const result = await fetchKanban(currentProject || undefined);
      setData(result);
      if (!currentProject && result.currentProject) {
        setCurrentProject(result.currentProject);
      }
      setError(null);
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Unknown error');
    } finally {
      setLoading(false);
    }
  }, [currentProject]);

  useEffect(() => {
    refresh();
  }, [refresh]);

  const addTask = useCallback(async (title: string, tags?: string[], dueAt?: string) => {
    if (!currentProject) return;
    try {
      await createTask({ project: currentProject, title, tags, dueAt });
      await refresh();
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Failed to create task');
    }
  }, [currentProject, refresh]);

  const moveTask = useCallback(async (taskId: number, newStatus: Task['status']) => {
    if (!currentProject) return;
    try {
      await updateTaskStatus(taskId, currentProject, newStatus);
      await refresh();
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Failed to move task');
    }
  }, [currentProject, refresh]);

  const removeTask = useCallback(async (taskId: number) => {
    if (!currentProject) return;
    try {
      await deleteTask(taskId, currentProject);
      await refresh();
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Failed to delete task');
    }
  }, [currentProject, refresh]);

  const editTask = useCallback(async (taskId: number, updates: { title?: string; tags?: string[]; dueAt?: string }) => {
    if (!currentProject) return;
    try {
      await updateTask(taskId, { project: currentProject, ...updates });
      await refresh();
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Failed to update task');
    }
  }, [currentProject, refresh]);

  const switchProject = useCallback((project: string) => {
    setCurrentProject(project);
  }, []);

  return {
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
  };
}
