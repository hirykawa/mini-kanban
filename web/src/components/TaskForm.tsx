import { useState } from 'react';
import { useTranslation } from '@/hooks/useTranslation';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { TagInput } from '@/components/ui/tag-input';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { AIAssistDialog } from './AIAssistDialog';
import { aiAssist, type AIAssistResponse } from '@/lib/api';

interface TaskFormProps {
  project: string;
  projects?: string[];
  availableTags?: string[];
  onSubmit: (title: string, tags?: string[], dueAt?: string, project?: string, body?: string) => void;
}

export function TaskForm({ project, projects, availableTags = [], onSubmit }: TaskFormProps) {
  const { t } = useTranslation();
  const [title, setTitle] = useState('');
  const [tags, setTags] = useState<string[]>([]);
  const [due, setDue] = useState('today');
  const [customDue, setCustomDue] = useState('');
  const [selectedProject, setSelectedProject] = useState(project);
  const [aiDialogOpen, setAiDialogOpen] = useState(false);
  const [pendingTitle, setPendingTitle] = useState('');
  const [initialQuestions, setInitialQuestions] = useState<AIAssistResponse | null>(null);
  const [enableAI, setEnableAI] = useState(true);
  const [isLoading, setIsLoading] = useState(false);

  const getDueDate = (preset: string): string => {
    const now = new Date();
    const today = new Date(now.getFullYear(), now.getMonth(), now.getDate(), 23, 59, 59);
    
    switch (preset) {
      case 'today':
        return today.toISOString().split('T')[0];
      case 'tomorrow': {
        const tomorrow = new Date(today);
        tomorrow.setDate(tomorrow.getDate() + 1);
        return tomorrow.toISOString().split('T')[0];
      }
      case 'this-week': {
        const daysUntilSunday = (7 - now.getDay()) % 7 || 7;
        const endOfWeek = new Date(today);
        endOfWeek.setDate(endOfWeek.getDate() + daysUntilSunday);
        return endOfWeek.toISOString().split('T')[0];
      }
      case 'this-month': {
        const endOfMonth = new Date(now.getFullYear(), now.getMonth() + 1, 0);
        return endOfMonth.toISOString().split('T')[0];
      }
      case 'next-month': {
        const endOfNextMonth = new Date(now.getFullYear(), now.getMonth() + 2, 0);
        return endOfNextMonth.toISOString().split('T')[0];
      }
      case '30m':
        return '30m';
      case '1h':
        return '1h';
      case '2h':
        return '2h';
      case 'custom':
        return customDue.trim() || '';
      default:
        return '';
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!title.trim()) return;

    const trimmedTitle = title.trim();

    if (enableAI) {
      setPendingTitle(trimmedTitle);
      setIsLoading(true);
      try {
        const targetProject = projects ? selectedProject : project;
        const response = await aiAssist({
          title: trimmedTitle,
          project: targetProject,
          existingTags: availableTags,
        });
        
        if (response.phase === 'skip' || !response.questions || response.questions.length === 0) {
          submitTask(trimmedTitle, tags, undefined);
          return;
        }
        
        setInitialQuestions(response);
        setAiDialogOpen(true);
      } catch {
        submitTask(trimmedTitle, tags, undefined);
      } finally {
        setIsLoading(false);
      }
      return;
    }

    submitTask(trimmedTitle, tags, undefined);
  };

  const submitTask = (taskTitle: string, taskTags: string[], body?: string) => {
    const tagList = taskTags.length > 0 ? taskTags : undefined;
    const dueAt = due !== 'none' ? getDueDate(due) : undefined;
    const targetProject = projects ? selectedProject : undefined;

    onSubmit(taskTitle, tagList, dueAt, targetProject, body);
    setTitle('');
    setTags([]);
  };

  const handleAIConfirm = (result: { title: string; body: string; tags: string[] }) => {
    const mergedTags = [...new Set([...tags, ...result.tags])];
    submitTask(result.title, mergedTags, result.body);
    setAiDialogOpen(false);
  };

  const handleAISkip = () => {
    if (pendingTitle) {
      submitTask(pendingTitle, tags, undefined);
    }
    setAiDialogOpen(false);
  };

  return (
    <>
      <form onSubmit={handleSubmit} className="flex gap-2 flex-wrap">
        <Input
          value={title}
          onChange={(e) => setTitle(e.target.value)}
          placeholder={t('newTaskPlaceholder')}
          className="flex-1 min-w-[200px]"
          required
        />
        <TagInput
          value={tags}
          onChange={setTags}
          suggestions={availableTags}
          placeholder={t('tagsPlaceholder')}
          className="w-[200px]"
        />
        <Select value={due} onValueChange={setDue}>
          <SelectTrigger className="w-[140px]">
            <SelectValue placeholder={t('dueDate')} />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="none">{t('noDueDate')}</SelectItem>
            <SelectItem value="30m">{t('inThirtyMinutes')}</SelectItem>
            <SelectItem value="1h">{t('inOneHour')}</SelectItem>
            <SelectItem value="2h">{t('inTwoHours')}</SelectItem>
            <SelectItem value="today">{t('today')}</SelectItem>
            <SelectItem value="tomorrow">{t('tomorrow')}</SelectItem>
            <SelectItem value="this-week">{t('thisWeek')}</SelectItem>
            <SelectItem value="this-month">{t('thisMonth')}</SelectItem>
            <SelectItem value="next-month">{t('nextMonth')}</SelectItem>
            <SelectItem value="custom">{t('customRelative')}</SelectItem>
          </SelectContent>
        </Select>
        {due === 'custom' && (
          <Input
            value={customDue}
            onChange={(e) => setCustomDue(e.target.value)}
            placeholder="2h, 30m, 1h30m..."
            className="w-[120px]"
          />
        )}
        {projects && projects.length > 1 && (
          <Select value={selectedProject} onValueChange={setSelectedProject}>
            <SelectTrigger className="w-[140px]">
              <SelectValue placeholder={t('project')} />
            </SelectTrigger>
            <SelectContent>
              {projects.map((p) => (
                <SelectItem key={p} value={p}>{p}</SelectItem>
              ))}
            </SelectContent>
          </Select>
        )}
        <label className="flex items-center gap-2 cursor-pointer">
          <input
            type="checkbox"
            checked={enableAI}
            onChange={(e) => setEnableAI(e.target.checked)}
            className="w-4 h-4 cursor-pointer"
          />
          <span className="text-sm">{t('enableAIAssist')}</span>
        </label>
        <Button type="submit" disabled={isLoading}>
          {isLoading ? (
            <>
              <span className="animate-spin rounded-full h-4 w-4 border-b-2 border-current"></span>
              {t('processing')}
            </>
          ) : (
            t('addTask')
          )}
        </Button>
      </form>

      <AIAssistDialog
        open={aiDialogOpen}
        onOpenChange={(open) => {
          setAiDialogOpen(open);
          if (!open) setInitialQuestions(null);
        }}
        title={pendingTitle}
        project={projects ? selectedProject : project}
        existingTags={availableTags}
        initialQuestions={initialQuestions}
        onConfirm={handleAIConfirm}
        onSkip={handleAISkip}
      />
    </>
  );
}
