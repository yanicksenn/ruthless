import React, { useState } from 'react';
import { X, Check, Search } from 'lucide-react';
import { Deck } from '../api/ruthless';
import { ConfigPublic_Limits } from '../api/config';

interface SessionCreationDialogProps {
  isOpen: boolean;
  onClose: () => void;
  onCreate: (deckIds: string[], name: string) => void;
  decks: Deck[];
  defaultName: string;
  limits: ConfigPublic_Limits | null;
}

export const SessionCreationDialog: React.FC<SessionCreationDialogProps> = ({
  isOpen,
  onClose,
  onCreate,
  decks,
  defaultName,
  limits
}) => {
  const [selectedIds, setSelectedIds] = useState<string[]>([]);
  const [sessionName, setSessionName] = useState('');
  const [filter, setFilter] = useState('');

  const maxNameLen = limits?.maxSessionNameLength ?? 64;
  const maxDecks = limits?.maxDecksPerSession ?? 16;
  
  const isNameInvalid = sessionName.length > maxNameLen;
  const isDeckLimitReached = selectedIds.length >= maxDecks;

  React.useEffect(() => {
    if (isOpen) {
      setSessionName(defaultName);
    }
  }, [isOpen, defaultName]);

  if (!isOpen) return null;

  const filteredDecks = decks.filter(d => 
    d.name.toLowerCase().includes(filter.toLowerCase())
  );

  const toggleDeck = (id: string) => {
    setSelectedIds(prev => {
      const isSelected = prev.includes(id);
      if (isSelected) return prev.filter(i => i !== id);
      if (prev.length < maxDecks) return [...prev, id];
      return prev;
    });
  };

  const handleCreate = () => {
    if (isNameInvalid || selectedIds.length === 0) return;
    onCreate(selectedIds, sessionName);
    setSelectedIds([]);
    setFilter('');
    setSessionName('');
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
          <div className="space-y-2">
            <div className="flex justify-between items-end px-4">
              <label className="text-[10px] font-black uppercase tracking-widest text-gray-500">
                Session Title
              </label>
              <span className={`text-[10px] font-black ${isNameInvalid ? 'text-red-400' : 'text-gray-500'}`}>
                {sessionName.length} / {maxNameLen}
              </span>
            </div>
            <input
              type="text"
              placeholder="Enter session title..."
              value={sessionName}
              onChange={(e) => setSessionName(e.target.value)}
              className={`w-full bg-white/5 border rounded-2xl py-3 px-6 text-white font-medium placeholder:text-gray-600 focus:outline-none focus:ring-2 transition-all font-bold text-lg ${
                isNameInvalid ? 'border-red-500/50 focus:ring-red-500/20' : 'border-white/10 focus:ring-primary/20'
              }`}
            />
            {isNameInvalid && (
              <p className="text-[10px] text-red-400 font-black uppercase tracking-wider pl-4 font-mono">
                Exceeds maximum length of {maxNameLen}
              </p>
            )}
          </div>

          <div className="space-y-2">
            <label className="text-[10px] font-black uppercase tracking-widest text-gray-500 ml-4">
              Filter Decks
            </label>
            <div className="relative group/search">
              <Search className="absolute left-4 top-1/2 -translate-y-1/2 text-gray-500 group-hover/search:text-primary transition-colors" size={18} />
              <input
                type="text"
                placeholder="Search for decks..."
                value={filter}
                onChange={(e) => setFilter(e.target.value)}
                className="w-full bg-white/5 border border-white/10 rounded-2xl py-3 pl-12 pr-4 text-white font-medium placeholder:text-gray-600 focus:outline-none focus:ring-2 focus:ring-primary/20 focus:border-primary/30 transition-all"
              />
            </div>
          </div>

          <div className="space-y-2">
            <div className="flex justify-between items-end px-4">
              <label className="text-[10px] font-black uppercase tracking-widest text-gray-500">
                Select Decks
              </label>
              <span className={`text-[10px] font-black ${isDeckLimitReached ? 'text-primary' : 'text-gray-500'}`}>
                {selectedIds.length} / {maxDecks}
              </span>
            </div>
            <div className="max-h-[300px] overflow-y-auto space-y-2 pr-2 custom-scrollbar">
              {filteredDecks.length === 0 ? (
                <div className="py-8 text-center bg-white/5 rounded-2xl border border-dashed border-white/10">
                  <p className="text-gray-500 font-bold italic">No decks found.</p>
                </div>
              ) : (
                filteredDecks.map(deck => {
                  const isSelected = selectedIds.includes(deck.id);
                  const isDisabled = !isSelected && isDeckLimitReached;
                  return (
                    <div
                      key={deck.id}
                      onClick={() => !isDisabled && toggleDeck(deck.id)}
                      className={`p-4 rounded-xl border transition-all flex justify-between items-center group/item ${
                        isSelected 
                          ? 'bg-primary/20 border-primary/30 text-white' 
                          : isDisabled
                            ? 'opacity-30 cursor-not-allowed border-white/5 bg-transparent'
                            : 'bg-white/5 border-white/10 text-gray-400 hover:bg-white/10 hover:border-white/20 cursor-pointer'
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
            {isDeckLimitReached && (
              <p className="text-[10px] text-primary font-black uppercase tracking-wider pl-4 font-mono">
                Maximum of {maxDecks} decks allowed
              </p>
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
              disabled={isNameInvalid || selectedIds.length === 0}
              className="flex-[2] px-6 py-4 rounded-2xl bg-secondary hover:bg-secondary/80 text-black font-black uppercase tracking-widest transition-all shadow-lg shadow-secondary/10 disabled:opacity-50 disabled:cursor-not-allowed"
            >
              Start Session {selectedIds.length > 0 && `(${selectedIds.length})`}
            </button>
          </div>
        </div>
      </div>
    </div>
  );
};
