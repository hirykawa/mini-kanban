import type { Project } from '@/lib/api';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';

interface HeaderProps {
  projects: Project[];
  currentProject: string;
  onProjectChange: (project: string) => void;
}

export function Header({ projects, currentProject, onProjectChange }: HeaderProps) {
  return (
    <header className="flex items-center justify-between pb-6 border-b mb-6">
      <h1 className="text-2xl font-bold bg-gradient-to-r from-blue-600 to-purple-600 bg-clip-text text-transparent">
        mini-kanban
      </h1>
      <div className="flex items-center gap-2">
        <label className="text-sm text-muted-foreground">Project:</label>
        <Select value={currentProject} onValueChange={onProjectChange}>
          <SelectTrigger className="w-[180px]">
            <SelectValue placeholder="Select project" />
          </SelectTrigger>
          <SelectContent>
            {projects.map((p) => (
              <SelectItem key={p.id} value={p.name}>
                {p.name}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      </div>
    </header>
  );
}
