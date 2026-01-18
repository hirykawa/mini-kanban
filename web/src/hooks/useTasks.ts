import { useState, useEffect, useCallback } from 'react';
import type { KanbanData, Task } from '@/lib/api';
import { fetchKanban, createTask, updateTaskStatus, deleteTask, updateTask } from '@/lib/api';

export function useTasks(initialProjects?: string[]) {
  const [data, setData] = useState<KanbanData | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [currentProjects, setCurrentProjects] = useState<string[]>(initialProjects || []);

  const refresh = useCallback(async () => {
    try {
      setLoading(true);
      const result = await fetchKanban(currentProjects.length > 0 ? currentProjects : undefined);
      setData(result);
      if (currentProjects.length === 0 && result.currentProjects && result.currentProjects.length > 0) {
        setCurrentProjects(result.currentProjects);
      }
      setError(null);
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Unknown error');
    } finally {
      setLoading(false);
    }
  }, [currentProjects]);

  useEffect(() => {
    refresh();
  }, [refresh]);

  // For single project operations, use the first selected project
  const primaryProject = currentProjects[0] || '';

  const addTask = useCallback(async (title: string, tags?: string[], dueAt?: string, project?: string, body?: string) => {
    const targetProject = project || primaryProject || data?.currentProject;
    if (!targetProject) return;
    try {
      await createTask({ project: targetProject, title, tags, dueAt, body });
      await refresh();
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Failed to create task');
    }
  }, [primaryProject, data?.currentProject, refresh]);

  const moveTask = useCallback(async (taskId: number, newStatus: Task['status'], project?: string) => {
    const targetProject = project || primaryProject;
    if (!targetProject) return;
    try {
      await updateTaskStatus(taskId, targetProject, newStatus);
      await refresh();
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Failed to move task');
    }
  }, [primaryProject, refresh]);

  const removeTask = useCallback(async (taskId: number, project?: string) => {
    const targetProject = project || primaryProject;
    if (!targetProject) return;
    try {
      await deleteTask(taskId, targetProject);
      await refresh();
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Failed to delete task');
    }
  }, [primaryProject, refresh]);

  const editTask = useCallback(async (taskId: number, updates: { title?: string; body?: string; tags?: string[]; dueAt?: string; project?: string }) => {
    const targetProject = updates.project || primaryProject;
    if (!targetProject) return;
    try {
      await updateTask(taskId, { project: targetProject, ...updates });
      await refresh();
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Failed to update task');
    }
  }, [primaryProject, refresh]);

  const switchProjects = useCallback((projects: string[]) => {
    setCurrentProjects(projects);
  }, []);

  const isMultiProject = currentProjects.length > 1;

  return {
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
  };
}
