import { useState, useRef, useEffect } from 'react';
import { cn } from '@/lib/utils';
import { Badge } from '@/components/ui/badge';
import { X, ChevronDown } from 'lucide-react';

interface TagInputProps {
  value: string[];
  onChange: (tags: string[]) => void;
  suggestions: string[];
  placeholder?: string;
  className?: string;
}

export function TagInput({
  value,
  onChange,
  suggestions,
  placeholder,
  className,
}: TagInputProps) {
  const [inputValue, setInputValue] = useState('');
  const [isOpen, setIsOpen] = useState(false);
  const [highlightedIndex, setHighlightedIndex] = useState(-1);
  const inputRef = useRef<HTMLInputElement>(null);
  const containerRef = useRef<HTMLDivElement>(null);

  // Filter suggestions: exclude already selected and filter by input
  const filteredSuggestions = suggestions.filter(
    (tag) =>
      !value.includes(tag) &&
      tag.toLowerCase().includes(inputValue.toLowerCase())
  );

  // Show "create new tag" option if input doesn't match any suggestion
  const showCreateOption =
    inputValue.trim() &&
    !suggestions.some(
      (tag) => tag.toLowerCase() === inputValue.trim().toLowerCase()
    ) &&
    !value.some(
      (tag) => tag.toLowerCase() === inputValue.trim().toLowerCase()
    );

  const allOptions = showCreateOption
    ? [...filteredSuggestions, `__create__:${inputValue.trim()}`]
    : filteredSuggestions;

  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (
        containerRef.current &&
        !containerRef.current.contains(event.target as Node)
      ) {
        setIsOpen(false);
      }
    };

    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, []);

  const addTag = (tag: string) => {
    const normalizedTag = tag.startsWith('__create__:')
      ? tag.replace('__create__:', '')
      : tag;
    if (normalizedTag && !value.includes(normalizedTag)) {
      onChange([...value, normalizedTag]);
    }
    setInputValue('');
    setHighlightedIndex(-1);
  };

  const removeTag = (tagToRemove: string) => {
    onChange(value.filter((tag) => tag !== tagToRemove));
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter') {
      e.preventDefault();
      if (highlightedIndex >= 0 && allOptions[highlightedIndex]) {
        addTag(allOptions[highlightedIndex]);
      } else if (inputValue.trim()) {
        addTag(inputValue.trim());
      }
    } else if (e.key === 'Backspace' && !inputValue && value.length > 0) {
      removeTag(value[value.length - 1]);
    } else if (e.key === 'ArrowDown') {
      e.preventDefault();
      setHighlightedIndex((prev) =>
        prev < allOptions.length - 1 ? prev + 1 : 0
      );
    } else if (e.key === 'ArrowUp') {
      e.preventDefault();
      setHighlightedIndex((prev) =>
        prev > 0 ? prev - 1 : allOptions.length - 1
      );
    } else if (e.key === 'Escape') {
      setIsOpen(false);
    } else if (e.key === ',') {
      e.preventDefault();
      if (inputValue.trim()) {
        addTag(inputValue.trim());
      }
    }
  };

  return (
    <div ref={containerRef} className={cn('relative', className)}>
      <div
        className={cn(
          'flex flex-wrap items-center gap-1 min-h-9 px-2 py-1 rounded-md border border-input bg-transparent shadow-xs transition-[color,box-shadow]',
          'focus-within:border-ring focus-within:ring-ring/50 focus-within:ring-[3px]',
          'dark:bg-input/30'
        )}
        onClick={() => inputRef.current?.focus()}
      >
        {value.map((tag) => (
          <Badge
            key={tag}
            variant="secondary"
            className="text-xs gap-1 pr-1"
          >
            {tag}
            <button
              type="button"
              onClick={(e) => {
                e.stopPropagation();
                removeTag(tag);
              }}
              className="hover:bg-muted rounded-sm"
            >
              <X className="h-3 w-3" />
            </button>
          </Badge>
        ))}
        <div className="flex-1 flex items-center min-w-[80px]">
          <input
            ref={inputRef}
            type="text"
            value={inputValue}
            onChange={(e) => {
              setInputValue(e.target.value);
              setIsOpen(true);
              setHighlightedIndex(-1);
            }}
            onFocus={() => setIsOpen(true)}
            onKeyDown={handleKeyDown}
            placeholder={value.length === 0 ? placeholder : ''}
            className="flex-1 bg-transparent outline-none text-sm placeholder:text-muted-foreground min-w-0"
          />
          <button
            type="button"
            onClick={() => setIsOpen(!isOpen)}
            className="p-1 hover:bg-muted rounded-sm"
          >
            <ChevronDown className={cn('h-4 w-4 text-muted-foreground transition-transform', isOpen && 'rotate-180')} />
          </button>
        </div>
      </div>

      {isOpen && allOptions.length > 0 && (
        <div className="absolute z-50 mt-1 w-full rounded-md border bg-popover text-popover-foreground shadow-md animate-in fade-in-0 zoom-in-95">
          <div className="max-h-[200px] overflow-y-auto p-1">
            {allOptions.map((option, index) => {
              const isCreateOption = option.startsWith('__create__:');
              const displayText = isCreateOption
                ? option.replace('__create__:', '')
                : option;

              return (
                <div
                  key={option}
                  className={cn(
                    'flex items-center px-2 py-1.5 text-sm rounded-sm cursor-pointer',
                    highlightedIndex === index && 'bg-accent text-accent-foreground',
                    'hover:bg-accent hover:text-accent-foreground'
                  )}
                  onClick={() => addTag(option)}
                  onMouseEnter={() => setHighlightedIndex(index)}
                >
                  {isCreateOption ? (
                    <span className="text-muted-foreground">
                      Create "<span className="text-foreground font-medium">{displayText}</span>"
                    </span>
                  ) : (
                    displayText
                  )}
                </div>
              );
            })}
          </div>
        </div>
      )}
    </div>
  );
}
