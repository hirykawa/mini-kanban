import { useState } from 'react';
import { useTranslation } from '@/hooks/useTranslation';
import type { Project } from '@/lib/api';
import { createProject } from '@/lib/api';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuCheckboxItem,
  DropdownMenuTrigger,
  DropdownMenuSeparator,
} from '@/components/ui/dropdown-menu';
import { Settings, Plus } from 'lucide-react';

interface HeaderProps {
  projects: Project[];
  currentProjects: string[];
  onProjectChange: (projects: string[]) => void;
  onSettingsClick: () => void;
  onProjectCreated?: () => void;
}

export function Header({ projects, currentProjects, onProjectChange, onSettingsClick, onProjectCreated }: HeaderProps) {
  const { t } = useTranslation();
  const [isCreating, setIsCreating] = useState(false);
  const [newProjectName, setNewProjectName] = useState('');
  const [error, setError] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);

  const toggleProject = (projectName: string) => {
    if (currentProjects.includes(projectName)) {
      // Don't allow deselecting if it's the only one
      if (currentProjects.length === 1) return;
      onProjectChange(currentProjects.filter(p => p !== projectName));
    } else {
      onProjectChange([...currentProjects, projectName]);
    }
  };

  const selectAll = () => {
    onProjectChange(projects.map(p => p.name));
  };

  const handleCreateProject = async () => {
    const name = newProjectName.trim();
    if (!name) return;

    setIsSubmitting(true);
    setError('');

    try {
      await createProject(name);
      setNewProjectName('');
      setIsCreating(false);
      onProjectCreated?.();
      onProjectChange([name]);
    } catch (err) {
      if (err instanceof Error && err.message.includes('already exists')) {
        setError(t('projectExists'));
      } else {
        setError(String(err));
      }
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter') {
      e.preventDefault();
      handleCreateProject();
    } else if (e.key === 'Escape') {
      setIsCreating(false);
      setNewProjectName('');
      setError('');
    }
  };

  const displayText = currentProjects.length === 1
    ? currentProjects[0]
    : currentProjects.length === projects.length
      ? t('allProjects')
      : t('nProjects', { count: currentProjects.length });

  return (
    <header className="flex items-center justify-between pb-6 border-b mb-6">
      <h1 className="text-2xl font-bold bg-gradient-to-r from-blue-600 to-purple-600 bg-clip-text text-transparent">
        mini-kanban
      </h1>
      <div className="flex items-center gap-2">
        <label className="text-sm text-muted-foreground">{t('projects')}</label>
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button variant="outline" className="w-[180px] justify-between">
              {displayText}
              <span className="ml-2 opacity-50">▼</span>
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent className="w-[180px]">
            {projects.length > 1 && (
              <DropdownMenuCheckboxItem
                checked={currentProjects.length === projects.length}
                onCheckedChange={selectAll}
              >
                {t('allProjects')}
              </DropdownMenuCheckboxItem>
            )}
            {projects.map((p) => (
              <DropdownMenuCheckboxItem
                key={p.id}
                checked={currentProjects.includes(p.name)}
                onCheckedChange={() => toggleProject(p.name)}
              >
                {p.name}
              </DropdownMenuCheckboxItem>
            ))}
            <DropdownMenuSeparator />
            {isCreating ? (
              <div className="px-2 py-1.5 space-y-2">
                <Input
                  autoFocus
                  value={newProjectName}
                  onChange={(e) => setNewProjectName(e.target.value)}
                  onKeyDown={handleKeyDown}
                  placeholder={t('projectNamePlaceholder')}
                  disabled={isSubmitting}
                  className="h-8"
                />
                {error && <p className="text-xs text-red-500">{error}</p>}
                <div className="flex gap-1">
                  <Button
                    size="sm"
                    onClick={handleCreateProject}
                    disabled={!newProjectName.trim() || isSubmitting}
                    className="flex-1 h-7"
                  >
                    {t('createProject')}
                  </Button>
                  <Button
                    size="sm"
                    variant="ghost"
                    onClick={() => {
                      setIsCreating(false);
                      setNewProjectName('');
                      setError('');
                    }}
                    className="h-7"
                  >
                    {t('cancel')}
                  </Button>
                </div>
              </div>
            ) : (
              <button
                className="flex items-center gap-2 w-full px-2 py-1.5 text-sm hover:bg-accent rounded-sm cursor-pointer"
                onClick={() => setIsCreating(true)}
              >
                <Plus className="h-4 w-4" />
                {t('newProject')}
              </button>
            )}
          </DropdownMenuContent>
        </DropdownMenu>
        <Button
          variant="ghost"
          size="icon"
          onClick={onSettingsClick}
          title="Settings"
        >
          <Settings className="h-5 w-5" />
        </Button>
      </div>
    </header>
  );
}
