import React, { useState } from 'react';
import { X, Check, Search } from 'lucide-react';
import { Deck } from '../api/ruthless';

interface SessionCreationDialogProps {
  isOpen: boolean;
  onClose: () => void;
  onCreate: (deckIds: string[]) => void;
  decks: Deck[];
}

export const SessionCreationDialog: React.FC<SessionCreationDialogProps> = ({
  isOpen,
  onClose,
  onCreate,
  decks
}) => {
  const [selectedIds, setSelectedIds] = useState<string[]>([]);
  const [filter, setFilter] = useState('');

  if (!isOpen) return null;

  const filteredDecks = decks.filter(d => 
    d.name.toLowerCase().includes(filter.toLowerCase())
  );

  const toggleDeck = (id: string) => {
    setSelectedIds(prev => 
      prev.includes(id) ? prev.filter(i => i !== id) : [...prev, id]
    );
  };

  const handleCreate = () => {
    onCreate(selectedIds);
    setSelectedIds([]);
    setFilter('');
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4">
      <div 
        className="absolute inset-0 bg-background/80 backdrop-blur-sm"
        onClick={onClose}
      />
      
      <div className="relative glass w-full max-w-xl rounded-[2.5rem] border border-white/5 shadow-2xl overflow-hidden animate-in fade-in zoom-in duration-300">
        <header className="p-8 pb-4 flex justify-between items-start">
          <div>
            <h2 className="text-3xl font-black text-white tracking-tight">Create Session</h2>
            <p className="text-gray-500 text-sm font-bold uppercase tracking-widest">Select decks for your chaos</p>
          </div>
          <button 
            onClick={onClose}
            className="p-2 hover:bg-white/5 rounded-xl text-gray-400 hover:text-white transition-all"
          >
            <X size={24} />
          </button>
        </header>

        <div className="px-8 pb-8 space-y-6">
          <div className="relative group/search">
            <Search className="absolute left-4 top-1/2 -translate-y-1/2 text-gray-500 group-hover/search:text-primary transition-colors" size={18} />
            <input
              type="text"
              placeholder="Filter decks..."
              value={filter}
              onChange={(e) => setFilter(e.target.value)}
              className="w-full bg-white/5 border border-white/10 rounded-2xl py-3 pl-12 pr-4 text-white font-medium placeholder:text-gray-600 focus:outline-none focus:ring-2 focus:ring-primary/20 focus:border-primary/30 transition-all"
            />
          </div>

          <div className="max-h-[300px] overflow-y-auto space-y-2 pr-2 custom-scrollbar">
            {filteredDecks.length === 0 ? (
              <div className="py-8 text-center bg-white/5 rounded-2xl border border-dashed border-white/10">
                <p className="text-gray-500 font-bold italic">No decks found.</p>
              </div>
            ) : (
              filteredDecks.map(deck => {
                const isSelected = selectedIds.includes(deck.id);
                return (
                  <div
                    key={deck.id}
                    onClick={() => toggleDeck(deck.id)}
                    className={`p-4 rounded-xl border transition-all cursor-pointer flex justify-between items-center group/item ${
                      isSelected 
                        ? 'bg-primary/20 border-primary/30 text-white' 
                        : 'bg-white/5 border-white/10 text-gray-400 hover:bg-white/10 hover:border-white/20'
                    }`}
                  >
                    <div>
                      <h4 className="font-bold">{deck.name}</h4>
                      <p className="text-[10px] font-black uppercase tracking-widest opacity-60">
                        {deck.cardIds?.length || 0} Cards
                      </p>
                    </div>
                    <div className={`w-6 h-6 rounded-lg border-2 flex items-center justify-center transition-all ${
                      isSelected 
                        ? 'bg-primary border-primary text-background' 
                        : 'border-white/10 group-hover/item:border-white/20'
                    }`}>
                      {isSelected && <Check size={14} strokeWidth={4} />}
                    </div>
                  </div>
                );
              })
            )}
          </div>

          <div className="flex gap-4 pt-4 border-t border-white/5">
            <button
              onClick={onClose}
              className="flex-1 px-6 py-4 rounded-2xl bg-white/5 hover:bg-white/10 text-gray-400 font-black uppercase tracking-widest transition-all"
            >
              Cancel
            </button>
            <button
              onClick={handleCreate}
              className="flex-[2] px-6 py-4 rounded-2xl bg-secondary hover:bg-secondary/80 text-black font-black uppercase tracking-widest transition-all shadow-lg shadow-secondary/10"
            >
              Start Session {selectedIds.length > 0 && `(${selectedIds.length})`}
            </button>
          </div>
        </div>
      </div>
    </div>
  );
};
