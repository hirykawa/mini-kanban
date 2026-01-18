import { useState, useEffect } from 'react';
import { useTranslation } from '@/hooks/useTranslation';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from '@/components/ui/dialog';
import { aiAssist, type AIAssistResponse } from '@/lib/api';

interface AIAssistDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  title: string;
  project: string;
  existingTags: string[];
  initialQuestions?: AIAssistResponse | null;
  onConfirm: (result: { title: string; body: string; tags: string[] }) => void;
  onSkip: () => void;
}

export function AIAssistDialog({
  open,
  onOpenChange,
  title,
  project,
  existingTags,
  initialQuestions,
  onConfirm,
  onSkip,
}: AIAssistDialogProps) {
  const { t } = useTranslation();
  const [phase, setPhase] = useState<'loading' | 'questions' | 'generating' | 'preview'>('loading');
  const [questions, setQuestions] = useState<string[]>([]);
  const [answers, setAnswers] = useState<string[]>([]);
  const [result, setResult] = useState<{ title: string; body: string; tags: string[] } | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    console.log('[AIAssistDialog] useEffect triggered', { open, title, initialQuestions });
    if (open && title) {
      // If we have pre-fetched questions, use them directly
      if (initialQuestions && initialQuestions.questions && initialQuestions.questions.length > 0) {
        console.log('[AIAssistDialog] using pre-fetched questions:', initialQuestions.questions);
        setQuestions(initialQuestions.questions);
        setAnswers(new Array(initialQuestions.questions.length).fill(''));
        setPhase('questions');
        setError(null);
        setResult(null);
      } else {
        // Fallback: fetch questions (shouldn't happen with new flow)
        console.log('[AIAssistDialog] no initialQuestions, calling fetchQuestions()');
        fetchQuestions();
      }
    }
  }, [open, title, initialQuestions]);

  const fetchQuestions = async () => {
    console.log('[AIAssistDialog] fetchQuestions called');
    setPhase('loading');
    setError(null);
    setQuestions([]);
    setAnswers([]);
    setResult(null);

    try {
      const response = await aiAssist({
        title,
        project,
        existingTags,
      });
      console.log('[AIAssistDialog] fetchQuestions response:', response);

      if (response.phase === 'skip') {
        // AI is disabled or title doesn't need assist - skip silently without opening dialog
        console.log('[AIAssistDialog] phase=skip, calling onSkip');
        onSkip();
        onOpenChange(false);
      } else if (response.questions && response.questions.length > 0) {
        console.log('[AIAssistDialog] setting questions:', response.questions);
        setQuestions(response.questions);
        setAnswers(new Array(response.questions.length).fill(''));
        setPhase('questions');
      } else {
        // No questions available, skip
        console.log('[AIAssistDialog] no questions, calling onSkip');
        onSkip();
      }
    } catch (err) {
      console.log('[AIAssistDialog] fetchQuestions error:', err);
      setError(t('aiAssistError'));
      setPhase('questions');
    }
  };

  const handleAnswerChange = (index: number, value: string) => {
    const newAnswers = [...answers];
    newAnswers[index] = value;
    setAnswers(newAnswers);
  };

  const handleSubmitAll = () => {
    generateResult(answers);
  };

  const generateResult = async (finalAnswers: string[]) => {
    setPhase('generating');
    setError(null);

    try {
      const response = await aiAssist({
        title,
        project,
        existingTags,
        answers: finalAnswers.filter(a => a.trim() !== ''),
      });

      if (response.phase === 'result' && response.title) {
        setResult({
          title: response.title,
          body: response.body || '',
          tags: response.tags || [],
        });
        setPhase('preview');
      } else {
        onSkip();
      }
    } catch (err) {
      setError(t('aiAssistError'));
      setPhase('preview');
    }
  };

  const handleConfirm = () => {
    if (result) {
      onConfirm(result);
    }
    onOpenChange(false);
  };

  const handleCancel = () => {
    onSkip();
    onOpenChange(false);
  };

  const hasAnyAnswer = answers.some(a => a.trim() !== '');

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[500px]">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <span>🤖</span>
            {t('aiAssistTitle')}
          </DialogTitle>
        </DialogHeader>

        <div className="py-4">
          {phase === 'loading' && (
            <div className="flex items-center justify-center py-8">
              <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
              <span className="ml-3">{t('aiAssistLoading')}</span>
            </div>
          )}

          {phase === 'questions' && questions.length > 0 && (
            <div className="space-y-4">
              <div className="text-sm text-muted-foreground">
                {t('aiAssistQuestionsDescription')}
              </div>
              {questions.map((question, index) => (
                <div key={index} className="space-y-2">
                  <div className="p-3 bg-muted rounded-lg">
                    <p className="font-medium text-sm">{question}</p>
                  </div>
                  <Input
                    value={answers[index]}
                    onChange={(e) => handleAnswerChange(index, e.target.value)}
                    placeholder={t('aiAssistAnswerPlaceholder')}
                  />
                </div>
              ))}
              <div className="flex gap-2 pt-2">
                <Button onClick={handleSubmitAll} disabled={!hasAnyAnswer}>
                  {t('generate')}
                </Button>
                <Button variant="ghost" onClick={handleCancel}>
                  {t('skip')}
                </Button>
              </div>
            </div>
          )}

          {phase === 'generating' && (
            <div className="flex items-center justify-center py-8">
              <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
              <span className="ml-3">{t('aiAssistGenerating')}</span>
            </div>
          )}

          {phase === 'preview' && result && (
            <div className="space-y-4">
              <div className="space-y-2">
                <label className="text-sm font-medium">{t('title')}</label>
                <div className="p-3 bg-muted rounded-lg">{result.title}</div>
              </div>
              {result.tags.length > 0 && (
                <div className="space-y-2">
                  <label className="text-sm font-medium">{t('tags')}</label>
                  <div className="flex flex-wrap gap-1">
                    {result.tags.map((tag, i) => (
                      <span key={i} className="px-2 py-1 bg-primary/10 text-primary rounded text-sm">
                        {tag}
                      </span>
                    ))}
                  </div>
                </div>
              )}
              {result.body && (
                <div className="space-y-2">
                  <label className="text-sm font-medium">{t('body')}</label>
                  <div className="p-3 bg-muted rounded-lg whitespace-pre-wrap text-sm max-h-[200px] overflow-y-auto">
                    {result.body}
                  </div>
                </div>
              )}
            </div>
          )}

          {error && (
            <div className="p-3 bg-destructive/10 text-destructive rounded-lg text-sm">
              {error}
            </div>
          )}
        </div>

        <DialogFooter>
          {phase === 'preview' && result && (
            <>
              <Button variant="ghost" onClick={handleCancel}>
                {t('useOriginal')}
              </Button>
              <Button onClick={handleConfirm}>
                {t('useAISuggestion')}
              </Button>
            </>
          )}
          {(phase === 'questions' || error) && (
            <Button variant="ghost" onClick={handleCancel}>
              {t('cancel')}
            </Button>
          )}
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
