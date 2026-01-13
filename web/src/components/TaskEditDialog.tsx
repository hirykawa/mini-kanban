import { useState } from 'react';
import type { Task } from '@/lib/api';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';

interface TaskEditDialogProps {
  task: Task;
  project: string;
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSave: (taskId: number, updates: { title?: string; tags?: string[]; dueAt?: string }) => void;
}

export function TaskEditDialog({ task, open, onOpenChange, onSave }: TaskEditDialogProps) {
  const [title, setTitle] = useState(task.title);
  const [tags, setTags] = useState(task.tags.join(', '));
  const [dueAt, setDueAt] = useState(() => {
    if (!task.dueAt) return '';
    const date = new Date(task.dueAt * 1000);
    return date.toISOString().split('T')[0];
  });

  const handleSave = () => {
    onSave(task.id, {
      title: title !== task.title ? title : undefined,
      tags: tags.split(',').map(t => t.trim()).filter(Boolean),
      dueAt: dueAt || 'none',
    });
    onOpenChange(false);
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[425px]">
        <DialogHeader>
          <DialogTitle>Edit Task #{task.id}</DialogTitle>
        </DialogHeader>
        <div className="grid gap-4 py-4">
          <div className="space-y-2">
            <label className="text-sm font-medium">Title</label>
            <Input
              value={title}
              onChange={(e) => setTitle(e.target.value)}
              placeholder="Task title"
            />
          </div>
          <div className="space-y-2">
            <label className="text-sm font-medium">Tags (comma separated)</label>
            <Input
              value={tags}
              onChange={(e) => setTags(e.target.value)}
              placeholder="tag1, tag2, tag3"
            />
          </div>
          <div className="space-y-2">
            <label className="text-sm font-medium">Due Date</label>
            <Input
              type="date"
              value={dueAt}
              onChange={(e) => setDueAt(e.target.value)}
            />
          </div>
        </div>
        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            Cancel
          </Button>
          <Button onClick={handleSave}>Save</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
