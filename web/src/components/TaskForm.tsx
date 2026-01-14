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

interface TaskFormProps {
  project: string;
  projects?: string[];
  availableTags?: string[];
  onSubmit: (title: string, tags?: string[], dueAt?: string, project?: string) => void;
}

export function TaskForm({ project, projects, availableTags = [], onSubmit }: TaskFormProps) {
  const { t } = useTranslation();
  const [title, setTitle] = useState('');
  const [tags, setTags] = useState<string[]>([]);
  const [due, setDue] = useState('today');
  const [selectedProject, setSelectedProject] = useState(project);

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
      default:
        return '';
    }
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!title.trim()) return;

    const tagList = tags.length > 0 ? tags : undefined;
    const dueAt = due !== 'none' ? getDueDate(due) : undefined;

    onSubmit(title.trim(), tagList, dueAt, projects ? selectedProject : undefined);
    setTitle('');
    setTags([]);
  };

  return (
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
          <SelectItem value="today">{t('today')}</SelectItem>
          <SelectItem value="tomorrow">{t('tomorrow')}</SelectItem>
          <SelectItem value="this-week">{t('thisWeek')}</SelectItem>
          <SelectItem value="this-month">{t('thisMonth')}</SelectItem>
          <SelectItem value="next-month">{t('nextMonth')}</SelectItem>
        </SelectContent>
      </Select>
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
      <Button type="submit">{t('addTask')}</Button>
    </form>
  );
}
