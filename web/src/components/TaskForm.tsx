import { useState } from 'react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';

interface TaskFormProps {
  project: string;
  onSubmit: (title: string, tags?: string[], dueAt?: string) => void;
}

export function TaskForm({ onSubmit }: TaskFormProps) {
  const [title, setTitle] = useState('');
  const [tags, setTags] = useState('');
  const [due, setDue] = useState('today');

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

    const tagList = tags ? tags.split(',').map(t => t.trim()).filter(Boolean) : undefined;
    const dueAt = due !== 'none' ? getDueDate(due) : undefined;

    onSubmit(title.trim(), tagList, dueAt);
    setTitle('');
    setTags('');
  };

  return (
    <form onSubmit={handleSubmit} className="flex gap-2 flex-wrap">
      <Input
        value={title}
        onChange={(e) => setTitle(e.target.value)}
        placeholder="New task title..."
        className="flex-1 min-w-[200px]"
        required
      />
      <Input
        value={tags}
        onChange={(e) => setTags(e.target.value)}
        placeholder="Tags (comma separated)"
        className="w-[180px]"
      />
      <Select value={due} onValueChange={setDue}>
        <SelectTrigger className="w-[140px]">
          <SelectValue placeholder="Due date" />
        </SelectTrigger>
        <SelectContent>
          <SelectItem value="none">No due date</SelectItem>
          <SelectItem value="today">Today</SelectItem>
          <SelectItem value="tomorrow">Tomorrow</SelectItem>
          <SelectItem value="this-week">This week</SelectItem>
          <SelectItem value="this-month">This month</SelectItem>
          <SelectItem value="next-month">Next month</SelectItem>
        </SelectContent>
      </Select>
      <Button type="submit">Add Task</Button>
    </form>
  );
}
