import { useState, useEffect } from 'react';
import { Button } from '@/components/ui/button';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Textarea } from '@/components/ui/textarea';

interface ConfigDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

interface Config {
  projectName: string;
  context: string;
  contextFile: string;
  configPath: string;
}

export function ConfigDialog({ open, onOpenChange }: ConfigDialogProps) {
  const [config, setConfig] = useState<Config>({
    projectName: '',
    context: '',
    contextFile: '',
    configPath: '',
  });
  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);

  useEffect(() => {
    if (open) {
      loadConfig();
    }
  }, [open]);

  const loadConfig = async () => {
    setLoading(true);
    try {
      const response = await fetch('/api/config');
      if (response.ok) {
        const data = await response.json();
        setConfig(data);
      }
    } catch (error) {
      console.error('Failed to load config:', error);
    } finally {
      setLoading(false);
    }
  };

  const saveConfig = async () => {
    setSaving(true);
    try {
      const response = await fetch('/api/config', {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          projectName: config.projectName,
          context: config.context,
          contextFile: config.contextFile,
        }),
      });
      if (response.ok) {
        onOpenChange(false);
      } else {
        console.error('Failed to save config');
      }
    } catch (error) {
      console.error('Failed to save config:', error);
    } finally {
      setSaving(false);
    }
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-2xl max-h-[80vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>Project Configuration</DialogTitle>
          <DialogDescription>
            Configure AI context for this project
            {config.configPath && (
              <div className="text-xs text-muted-foreground mt-1">
                {config.configPath}
              </div>
            )}
          </DialogDescription>
        </DialogHeader>

        {loading ? (
          <div className="py-8 text-center text-muted-foreground">Loading...</div>
        ) : (
          <div className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="projectName">Project Name</Label>
              <Input
                id="projectName"
                value={config.projectName}
                onChange={(e) =>
                  setConfig({ ...config, projectName: e.target.value })
                }
                placeholder="my-project"
              />
            </div>

            <div className="space-y-2">
              <Label htmlFor="contextFile">Context File Path</Label>
              <Input
                id="contextFile"
                value={config.contextFile}
                onChange={(e) =>
                  setConfig({ ...config, contextFile: e.target.value })
                }
                placeholder="README.md"
              />
              <p className="text-xs text-muted-foreground">
                Relative path to a file containing AI context (e.g., README.md)
              </p>
            </div>

            <div className="space-y-2">
              <Label htmlFor="context">Inline Context</Label>
              <Textarea
                id="context"
                value={config.context}
                onChange={(e) =>
                  setConfig({ ...config, context: e.target.value })
                }
                placeholder="Enter project context for AI assistance..."
                rows={8}
                className="font-mono text-sm"
              />
              <p className="text-xs text-muted-foreground">
                Direct context text (used if context file is not found)
              </p>
            </div>

            <div className="flex justify-end gap-2 pt-4">
              <Button
                variant="outline"
                onClick={() => onOpenChange(false)}
                disabled={saving}
              >
                Cancel
              </Button>
              <Button onClick={saveConfig} disabled={saving}>
                {saving ? 'Saving...' : 'Save'}
              </Button>
            </div>
          </div>
        )}
      </DialogContent>
    </Dialog>
  );
}
