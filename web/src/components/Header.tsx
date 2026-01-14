import { useTranslation } from '@/hooks/useTranslation';
import type { Project } from '@/lib/api';
import { Button } from '@/components/ui/button';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuCheckboxItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';

interface HeaderProps {
  projects: Project[];
  currentProjects: string[];
  onProjectChange: (projects: string[]) => void;
}

export function Header({ projects, currentProjects, onProjectChange }: HeaderProps) {
  const { t } = useTranslation();
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
          </DropdownMenuContent>
        </DropdownMenu>
      </div>
    </header>
  );
}
