import { useState, useMemo, useEffect } from 'react';
import { useTranslation } from '@/hooks/useTranslation';
import type { KanbanData, Task } from '@/lib/api';
import { KanbanColumn } from './KanbanColumn';
import { TaskForm } from './TaskForm';
import { TagInput } from './ui/tag-input';
import { Badge } from './ui/badge';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from './ui/select';

export type DoneFilterType = 'all' | 'today' | 'week' | 'month';

const FILTER_STORAGE_KEY = 'kanban-filters';

interface FilterState {
  tags: string[];
  doneFilter: DoneFilterType;
}

function loadFilters(): FilterState {
  try {
    const saved = localStorage.getItem(FILTER_STORAGE_KEY);
    if (saved) {
      const parsed = JSON.parse(saved);
      return {
        tags: Array.isArray(parsed.tags) ? parsed.tags : [],
        doneFilter: ['all', 'today', 'week', 'month'].includes(parsed.doneFilter) 
          ? parsed.doneFilter 
          : 'today',
      };
    }
  } catch {
  }
  return { tags: [], doneFilter: 'today' };
}

function saveFilters(filters: FilterState) {
  try {
    localStorage.setItem(FILTER_STORAGE_KEY, JSON.stringify(filters));
  } catch {
  }
}

function getStartOfDay(date: Date): number {
  const d = new Date(date);
  d.setHours(0, 0, 0, 0);
  return Math.floor(d.getTime() / 1000);
}

function getStartOfWeek(date: Date): number {
  const d = new Date(date);
  const day = d.getDay();
  const diff = d.getDate() - day + (day === 0 ? -6 : 1);
  d.setDate(diff);
  d.setHours(0, 0, 0, 0);
  return Math.floor(d.getTime() / 1000);
}

function getStartOfMonth(date: Date): number {
  const d = new Date(date);
  d.setDate(1);
  d.setHours(0, 0, 0, 0);
  return Math.floor(d.getTime() / 1000);
}

interface KanbanBoardProps {
  data: KanbanData;
  isMultiProject: boolean;
  onAddTask: (title: string, tags?: string[], dueAt?: string, project?: string, body?: string) => void;
  onMoveTask: (taskId: number, newStatus: Task['status'], project: string) => void;
  onDeleteTask: (taskId: number, project: string) => void;
  onEditTask: (taskId: number, updates: { title?: string; body?: string; tags?: string[]; dueAt?: string; project?: string }) => void;
}

export function KanbanBoard({ data, isMultiProject, onAddTask, onMoveTask, onDeleteTask, onEditTask }: KanbanBoardProps) {
  const { t } = useTranslation();
  const [filterTags, setFilterTags] = useState<string[]>(() => loadFilters().tags);
  const [doneFilter, setDoneFilter] = useState<DoneFilterType>(() => loadFilters().doneFilter);

  useEffect(() => {
    saveFilters({ tags: filterTags, doneFilter });
  }, [filterTags, doneFilter]);

  const filteredData = useMemo(() => {
    const filterByTags = (tasks: Task[]) => {
      if (filterTags.length === 0) return tasks;
      return tasks.filter(task => 
        filterTags.every(tag => task.tags.includes(tag))
      );
    };

    const filterDoneTasks = (tasks: Task[]) => {
      let filtered = filterByTags(tasks);
      
      if (doneFilter === 'all') {
        return filtered;
      }
      
      const now = new Date();
      let startTime: number;
      
      switch (doneFilter) {
        case 'today':
          startTime = getStartOfDay(now);
          break;
        case 'week':
          startTime = getStartOfWeek(now);
          break;
        case 'month':
          startTime = getStartOfMonth(now);
          break;
        default:
          return filtered;
      }
      
      return filtered.filter(task => task.updatedAt >= startTime);
    };

    return {
      todoTasks: filterByTags(data.todoTasks || []),
      doingTasks: filterByTags(data.doingTasks || []),
      reviewTasks: filterByTags(data.reviewTasks || []),
      doneTasks: filterDoneTasks(data.doneTasks || []),
    };
  }, [data, filterTags, doneFilter]);

  return (
    <div className="space-y-6">
      <TaskForm
        project={data.currentProject}
        projects={isMultiProject ? data.currentProjects : undefined}
        availableTags={data.tags}
        onSubmit={onAddTask}
      />
      
      <div className="flex items-center gap-4 flex-wrap">
        <span className="text-sm font-medium text-muted-foreground">{t('filters')}:</span>
        
        {data.tags && data.tags.length > 0 && (
          <div className="flex items-center gap-2">
            <span className="text-sm text-muted-foreground">{t('filterByTags')}:</span>
            <TagInput
              value={filterTags}
              onChange={setFilterTags}
              suggestions={data.tags}
              placeholder={t('selectTags')}
              className="w-[250px]"
            />
            {filterTags.length > 0 && (
              <Badge variant="secondary" className="text-xs">
                {t('filteringCount', { count: filterTags.length })}
              </Badge>
            )}
          </div>
        )}
        
        <div className="flex items-center gap-2">
          <span className="text-sm text-muted-foreground">{t('doneFilter')}:</span>
          <Select value={doneFilter} onValueChange={(v) => setDoneFilter(v as DoneFilterType)}>
            <SelectTrigger className="w-[140px]">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="today">{t('doneFilterToday')}</SelectItem>
              <SelectItem value="week">{t('doneFilterThisWeek')}</SelectItem>
              <SelectItem value="month">{t('doneFilterThisMonth')}</SelectItem>
              <SelectItem value="all">{t('doneFilterAll')}</SelectItem>
            </SelectContent>
          </Select>
        </div>
      </div>
      
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
        <KanbanColumn
          title={t('todo')}
          status="todo"
          tasks={filteredData.todoTasks}
          showProject={isMultiProject}
          availableTags={data.tags}
          onMove={onMoveTask}
          onDelete={onDeleteTask}
          onEdit={onEditTask}
        />
        <KanbanColumn
          title={t('doing')}
          status="doing"
          tasks={filteredData.doingTasks}
          showProject={isMultiProject}
          availableTags={data.tags}
          onMove={onMoveTask}
          onDelete={onDeleteTask}
          onEdit={onEditTask}
        />
        <KanbanColumn
          title={t('preview')}
          status="review"
          tasks={filteredData.reviewTasks}
          showProject={isMultiProject}
          availableTags={data.tags}
          onMove={onMoveTask}
          onDelete={onDeleteTask}
          onEdit={onEditTask}
        />
        <KanbanColumn
          title={t('done')}
          status="done"
          tasks={filteredData.doneTasks}
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
