// Kanban drag and drop functionality
let draggedCard = null;

function handleDragStart(e) {
    draggedCard = e.target;
    e.target.classList.add('dragging');
    e.dataTransfer.effectAllowed = 'move';
    e.dataTransfer.setData('text/plain', e.target.dataset.id);
    
    // Add visual feedback
    setTimeout(() => {
        e.target.style.opacity = '0.5';
    }, 0);
}

function handleDragEnd(e) {
    e.target.classList.remove('dragging');
    e.target.style.opacity = '1';
    draggedCard = null;
    
    // Remove all drop indicators
    document.querySelectorAll('.drag-over').forEach(el => el.classList.remove('drag-over'));
}

function handleDragOver(e) {
    e.preventDefault();
    e.dataTransfer.dropEffect = 'move';
    
    const cardsContainer = e.currentTarget;
    cardsContainer.classList.add('drag-over');
}

function handleDragLeave(e) {
    // Only remove if we're actually leaving the container
    if (!e.currentTarget.contains(e.relatedTarget)) {
        e.currentTarget.classList.remove('drag-over');
    }
}

async function handleDrop(e, newStatus) {
    e.preventDefault();
    e.currentTarget.classList.remove('drag-over');
    
    const taskId = e.dataTransfer.getData('text/plain');
    if (!taskId) return;
    
    // Get current project from the page
    const projectSelect = document.getElementById('project');
    const project = projectSelect ? projectSelect.value : 'default';
    
    // Send status update to server
    try {
        const formData = new FormData();
        formData.append('project', project);
        formData.append('status', newStatus);
        
        const response = await fetch(`/tasks/${taskId}/status`, {
            method: 'POST',
            body: formData
        });
        
        if (response.ok) {
            // Trigger HTMX refresh
            document.body.dispatchEvent(new CustomEvent('taskUpdated'));
        } else {
            console.error('Failed to update task status');
        }
    } catch (error) {
        console.error('Error updating task status:', error);
    }
}

// Re-initialize HTMX after page loads
document.addEventListener('DOMContentLoaded', function() {
    htmx.process(document.body);
});

// Re-process HTMX after swaps
document.body.addEventListener('htmx:afterSwap', function(e) {
    htmx.process(e.target);
});
