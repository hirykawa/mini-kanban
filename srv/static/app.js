// Mini Kanban App
let boardData = null;
let draggedCard = null;

async function loadBoard() {
    const response = await fetch('/api/board');
    boardData = await response.json();
    renderBoard();
}

function renderBoard() {
    const board = document.getElementById('board');
    board.innerHTML = '';
    
    boardData.columns.forEach(column => {
        board.appendChild(createColumnElement(column));
    });
    
    // Add column button
    const addColumnBtn = document.createElement('div');
    addColumnBtn.className = 'add-column';
    addColumnBtn.innerHTML = '+ Add column';
    addColumnBtn.onclick = () => showAddColumnForm();
    board.appendChild(addColumnBtn);
}

function createColumnElement(column) {
    const el = document.createElement('div');
    el.className = 'column';
    el.dataset.columnId = column.id;
    
    el.innerHTML = `
        <div class="column-header">
            <h2>${escapeHtml(column.name)}</h2>
            <button class="delete-column" onclick="deleteColumn(${column.id})">&times;</button>
        </div>
        <div class="column-cards" ondragover="handleDragOver(event)" ondragleave="handleDragLeave(event)" ondrop="handleDrop(event, ${column.id})">
            ${column.cards.map(card => createCardHTML(card)).join('')}
        </div>
        <button class="add-card-btn" onclick="showAddCardForm(${column.id})">+ Add a card</button>
    `;
    
    return el;
}

function createCardHTML(card) {
    return `
        <div class="card" draggable="true" data-card-id="${card.id}" 
             ondragstart="handleDragStart(event)" ondragend="handleDragEnd(event)">
            <div class="card-title">${escapeHtml(card.title)}</div>
            ${card.description ? `<div class="card-description">${escapeHtml(card.description)}</div>` : ''}
            <div class="card-actions">
                <button onclick="editCard(${card.id})">Edit</button>
                <button onclick="deleteCard(${card.id})">Delete</button>
            </div>
        </div>
    `;
}

function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

// Drag and Drop
function handleDragStart(e) {
    draggedCard = e.target;
    e.target.classList.add('dragging');
    e.dataTransfer.effectAllowed = 'move';
    e.dataTransfer.setData('text/plain', e.target.dataset.cardId);
}

function handleDragEnd(e) {
    e.target.classList.remove('dragging');
    draggedCard = null;
    document.querySelectorAll('.drag-over').forEach(el => el.classList.remove('drag-over'));
    document.querySelectorAll('.drop-placeholder').forEach(el => el.remove());
}

function handleDragOver(e) {
    e.preventDefault();
    e.dataTransfer.dropEffect = 'move';
    const cardsContainer = e.currentTarget;
    cardsContainer.classList.add('drag-over');
}

function handleDragLeave(e) {
    e.currentTarget.classList.remove('drag-over');
}

async function handleDrop(e, columnId) {
    e.preventDefault();
    e.currentTarget.classList.remove('drag-over');
    
    const cardId = parseInt(e.dataTransfer.getData('text/plain'));
    
    // Find the position to insert at
    const cardsContainer = e.currentTarget;
    const cards = Array.from(cardsContainer.querySelectorAll('.card:not(.dragging)'));
    let position = cards.length;
    
    // Calculate position based on mouse Y
    for (let i = 0; i < cards.length; i++) {
        const rect = cards[i].getBoundingClientRect();
        if (e.clientY < rect.top + rect.height / 2) {
            position = i;
            break;
        }
    }
    
    await fetch(`/api/cards/${cardId}/move`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ column_id: columnId, position: position })
    });
    
    await loadBoard();
}

// Add Card
function showAddCardForm(columnId) {
    const column = document.querySelector(`[data-column-id="${columnId}"]`);
    const cardsContainer = column.querySelector('.column-cards');
    const addBtn = column.querySelector('.add-card-btn');
    
    addBtn.style.display = 'none';
    
    const form = document.createElement('div');
    form.className = 'card-form';
    form.innerHTML = `
        <textarea placeholder="Enter a title for this card..." rows="3" autofocus></textarea>
        <div class="card-form-actions">
            <button class="save-btn" onclick="addCard(${columnId}, this)">Add Card</button>
            <button class="cancel-btn" onclick="cancelAddCard(${columnId}, this)">Cancel</button>
        </div>
    `;
    
    cardsContainer.appendChild(form);
    form.querySelector('textarea').focus();
    
    // Handle Enter key
    form.querySelector('textarea').addEventListener('keydown', (e) => {
        if (e.key === 'Enter' && !e.shiftKey) {
            e.preventDefault();
            addCard(columnId, form.querySelector('.save-btn'));
        }
        if (e.key === 'Escape') {
            cancelAddCard(columnId, form.querySelector('.cancel-btn'));
        }
    });
}

async function addCard(columnId, btn) {
    const form = btn.closest('.card-form');
    const title = form.querySelector('textarea').value.trim();
    
    if (!title) return;
    
    await fetch('/api/cards', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ column_id: columnId, title: title, description: '' })
    });
    
    await loadBoard();
}

function cancelAddCard(columnId, btn) {
    const form = btn.closest('.card-form');
    form.remove();
    const column = document.querySelector(`[data-column-id="${columnId}"]`);
    column.querySelector('.add-card-btn').style.display = '';
}

// Edit Card
function editCard(cardId) {
    // Find card data
    let card = null;
    for (const col of boardData.columns) {
        card = col.cards.find(c => c.id === cardId);
        if (card) break;
    }
    if (!card) return;
    
    const modal = document.createElement('div');
    modal.className = 'modal-overlay';
    modal.innerHTML = `
        <div class="modal">
            <h3>Edit Card</h3>
            <input type="text" id="edit-title" value="${escapeHtml(card.title)}" placeholder="Title">
            <textarea id="edit-description" rows="4" placeholder="Description">${escapeHtml(card.description)}</textarea>
            <div class="modal-actions">
                <button class="cancel-btn" onclick="this.closest('.modal-overlay').remove()">Cancel</button>
                <button class="save-btn" onclick="saveCardEdit(${cardId})">Save</button>
            </div>
        </div>
    `;
    
    document.body.appendChild(modal);
    modal.querySelector('#edit-title').focus();
    
    modal.addEventListener('click', (e) => {
        if (e.target === modal) modal.remove();
    });
}

async function saveCardEdit(cardId) {
    const title = document.getElementById('edit-title').value.trim();
    const description = document.getElementById('edit-description').value.trim();
    
    if (!title) return;
    
    await fetch(`/api/cards/${cardId}`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ title, description })
    });
    
    document.querySelector('.modal-overlay').remove();
    await loadBoard();
}

// Delete Card
async function deleteCard(cardId) {
    if (!confirm('Delete this card?')) return;
    
    await fetch(`/api/cards/${cardId}`, { method: 'DELETE' });
    await loadBoard();
}

// Add Column
function showAddColumnForm() {
    const modal = document.createElement('div');
    modal.className = 'modal-overlay';
    modal.innerHTML = `
        <div class="modal">
            <h3>Add Column</h3>
            <input type="text" id="column-name" placeholder="Column name" autofocus>
            <div class="modal-actions">
                <button class="cancel-btn" onclick="this.closest('.modal-overlay').remove()">Cancel</button>
                <button class="save-btn" onclick="addColumn()">Add</button>
            </div>
        </div>
    `;
    
    document.body.appendChild(modal);
    modal.querySelector('#column-name').focus();
    
    modal.querySelector('#column-name').addEventListener('keydown', (e) => {
        if (e.key === 'Enter') addColumn();
        if (e.key === 'Escape') modal.remove();
    });
    
    modal.addEventListener('click', (e) => {
        if (e.target === modal) modal.remove();
    });
}

async function addColumn() {
    const name = document.getElementById('column-name').value.trim();
    if (!name) return;
    
    await fetch('/api/columns', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ name })
    });
    
    document.querySelector('.modal-overlay').remove();
    await loadBoard();
}

// Delete Column
async function deleteColumn(columnId) {
    if (!confirm('Delete this column and all its cards?')) return;
    
    await fetch(`/api/columns/${columnId}`, { method: 'DELETE' });
    await loadBoard();
}

// Initialize
document.addEventListener('DOMContentLoaded', loadBoard);
