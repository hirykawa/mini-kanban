import { useState } from 'react';
import ReactMarkdown from 'react-markdown';
import { useTranslation } from '@/hooks/useTranslation';
import type { Task, AIAssistResponse } from '@/lib/api';
import { aiAssist } from '@/lib/api';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { TagInput } from '@/components/ui/tag-input';
import { cn } from "@/lib/utils";
import { AIAssistDialog } from './AIAssistDialog';

interface TaskEditDialogProps {
  task: Task;
  project: string;
  availableTags?: string[];
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSave: (taskId: number, updates: { title?: string; body?: string; tags?: string[]; dueAt?: string }) => void;
}

export function TaskEditDialog({ task, project, availableTags = [], open, onOpenChange, onSave }: TaskEditDialogProps) {
  const { t } = useTranslation();
  const [title, setTitle] = useState(task.title);
  const [body, setBody] = useState(task.body || '');
  const [tags, setTags] = useState<string[]>(task.tags);
  const [dueAt, setDueAt] = useState(() => {
    if (!task.dueAt) return '';
    const date = new Date(task.dueAt * 1000);
    return date.toISOString().split('T')[0];
  });
  const [isPreview, setIsPreview] = useState(false);
  const [aiDialogOpen, setAiDialogOpen] = useState(false);
  const [pendingTitle, setPendingTitle] = useState('');
  const [initialQuestions, setInitialQuestions] = useState<AIAssistResponse | null>(null);

  const handleSave = () => {
    onSave(task.id, {
      title: title !== task.title ? title : undefined,
      body: body !== (task.body || '') ? body : undefined,
      tags: tags,
      dueAt: dueAt || 'none',
    });
    onOpenChange(false);
  };

  const handleAIAssist = async () => {
    if (!title.trim()) return;
    const trimmedTitle = title.trim();
    setPendingTitle(trimmedTitle);
    
    try {
      const response = await aiAssist({
        title: trimmedTitle,
        project,
        existingTags: availableTags,
      });
      
      // If AI says skip, don't open dialog
      if (response.phase === 'skip' || !response.questions || response.questions.length === 0) {
        return;
      }
      
      // Has questions, open dialog with pre-fetched data
      setInitialQuestions(response);
      setAiDialogOpen(true);
    } catch {
      // On error, do nothing
    }
  };

  const handleAIConfirm = (result: { title: string; body: string; tags: string[] }) => {
    setTitle(result.title);
    setBody(result.body);
    // Merge AI suggested tags with existing tags
    const mergedTags = [...new Set([...tags, ...result.tags])];
    setTags(mergedTags);
    setAiDialogOpen(false);
  };

  const handleAISkip = () => {
    setAiDialogOpen(false);
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[600px] max-h-[85vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>{t('editTaskTitle', { id: task.id })}</DialogTitle>
        </DialogHeader>
        <div className="grid gap-4 py-4">
          <div className="space-y-2">
            <div className="flex justify-between items-center">
              <label className="text-sm font-medium">{t('title')}</label>
              <Button 
                variant="ghost" 
                size="sm" 
                onClick={handleAIAssist}
                className="h-6 text-xs"
              >
                🤖 {t('enableAIAssist')}
              </Button>
            </div>
            <Input
              value={title}
              onChange={(e) => setTitle(e.target.value)}
              placeholder={t('taskTitlePlaceholder')}
            />
          </div>
          
          <div className="space-y-2">
            <div className="flex justify-between items-center">
              <label className="text-sm font-medium">{t('description')}</label>
              <Button 
                variant="ghost" 
                size="sm" 
                onClick={() => setIsPreview(!isPreview)}
                className="h-6 text-xs"
              >
                {isPreview ? t('edit') : t('preview')}
              </Button>
            </div>
             {isPreview ? (
              <div className="min-h-[150px] p-3 border rounded-md bg-transparent text-sm overflow-y-auto max-h-[300px]">
                 {body ? (
                   <ReactMarkdown
                    components={{
                      h1: (props) => <h1 className="text-xl font-bold mb-2 mt-4 first:mt-0" {...props} />,
                      h2: (props) => <h2 className="text-lg font-bold mb-2 mt-4" {...props} />,
                      h3: (props) => <h3 className="text-base font-bold mb-2 mt-2" {...props} />,
                      p: (props) => <p className="mb-2 leading-relaxed" {...props} />,
                      ul: (props) => <ul className="list-disc list-inside mb-2" {...props} />,
                      ol: (props) => <ol className="list-decimal list-inside mb-2" {...props} />,
                      code: (props) => <code className="bg-muted px-1.5 py-0.5 rounded text-xs font-mono" {...props} />,
                      pre: (props) => <pre className="bg-muted p-2 rounded mb-2 overflow-x-auto text-xs" {...props} />,
                      blockquote: (props) => <blockquote className="border-l-4 border-muted pl-2 text-muted-foreground italic mb-2" {...props} />,
                      a: (props) => <a className="text-blue-500 hover:underline" target="_blank" rel="noopener noreferrer" {...props} />,
                    }}
                   >
                     {body}
                   </ReactMarkdown>
                 ) : (
                   <span className="text-muted-foreground italic">{t('noDescription')}</span>
                 )}
              </div>
             ) : (
              <textarea
                value={body}
                onChange={(e) => setBody(e.target.value)}
                placeholder={t('descriptionPlaceholder')}
                className={cn(
                  "file:text-foreground placeholder:text-muted-foreground selection:bg-primary selection:text-primary-foreground dark:bg-input/30 border-input w-full min-w-0 rounded-md border bg-transparent px-3 py-2 text-base shadow-xs transition-[color,box-shadow] outline-none disabled:pointer-events-none disabled:cursor-not-allowed disabled:opacity-50 md:text-sm",
                  "focus-visible:border-ring focus-visible:ring-ring/50 focus-visible:ring-[3px]",
                  "min-h-[150px] resize-y"
                )}
              />
            )}
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2">
              <label className="text-sm font-medium">{t('tagsSimple')}</label>
              <TagInput
                value={tags}
                onChange={setTags}
                suggestions={availableTags}
                placeholder={t('tagsExamplePlaceholder')}
              />
            </div>
            <div className="space-y-2">
              <label className="text-sm font-medium">{t('dueDate')}</label>
              <Input
                type="date"
                value={dueAt}
                onChange={(e) => setDueAt(e.target.value)}
              />
            </div>
          </div>
        </div>
        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            {t('cancel')}
          </Button>
          <Button onClick={handleSave}>{t('save')}</Button>
        </DialogFooter>
      </DialogContent>

      <AIAssistDialog
        open={aiDialogOpen}
        onOpenChange={(open) => {
          setAiDialogOpen(open);
          if (!open) setInitialQuestions(null);
        }}
        title={pendingTitle}
        project={project}
        existingTags={availableTags}
        initialQuestions={initialQuestions}
        onConfirm={handleAIConfirm}
        onSkip={handleAISkip}
      />
    </Dialog>
  );
}
