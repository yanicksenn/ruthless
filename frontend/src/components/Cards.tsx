import React, { useState, useEffect } from 'react';
import { useAuth } from '../context/AuthContext';
import { Plus, Trash2, LogOut, Search, ArrowUpDown, ChevronDown, FolderPlus, Check, X, Layers } from 'lucide-react';
import { cardClient, deckClient, createOptions } from '../api/client';
import { Card, CardOrderField, Deck } from '../api/ruthless';
import { CreationDialog } from './CreationDialog';

export const Cards: React.FC = () => {
  const { token, user, logout, limits } = useAuth();
  const [cards, setCards] = useState<Card[]>([]);
  const [decks, setDecks] = useState<Deck[]>([]);
  const [loading, setLoading] = useState(true);
  const [isDialogOpen, setIsDialogOpen] = useState(false);
  const [pageNumber, setPageNumber] = useState(1);
  const [totalCount, setTotalCount] = useState(0);
  const [filter, setFilter] = useState('');
  const [sortField, setSortField] = useState<CardOrderField>(CardOrderField.CREATED_AT);
  const [descending, setDescending] = useState(true);
  const [activeCardIdDropdown, setActiveCardIdDropdown] = useState<string | null>(null);
  const pageSize = 12;

  const fetchData = async () => {
    setLoading(true);
    try {
      const [cardsRes, decksRes] = await Promise.all([
        cardClient.listCards({ 
          pageSize, 
          pageNumber, 
          ids: [], 
          filter,
          orderBy: {
            field: sortField,
            descending
          }
        }, createOptions(token)),
        deckClient.listDecks({}, createOptions(token))
      ]);
      
      setCards(cardsRes.response.cards || []);
      setTotalCount(cardsRes.response.totalCount);
      setDecks(decksRes.response.decks || []);
    } catch (err) {
      console.error('Failed to fetch data:', err);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchData();
  }, [token, pageNumber, filter, sortField, descending]);

  // Reset to first page when filters change
  useEffect(() => {
    setPageNumber(1);
  }, [filter, sortField, descending]);

  const totalPages = Math.ceil(totalCount / pageSize);

  const handleOpenCardDialog = () => {
    setIsDialogOpen(true);
  };

  const handleCreate = async (value: string) => {
    try {
      await cardClient.createCard({ text: value }, createOptions(token));
      fetchData();
    } catch (err) {
      alert(`Failed to create card`);
    }
  };

  const handleDelete = async (id: string) => {
    if (!confirm('Are you sure you want to delete this card?')) return;
    try {
      await cardClient.deleteCard({ id }, createOptions(token));
      fetchData();
    } catch (err: any) {
      alert(`Failed to delete card: ${err.message || err}`);
    }
  };

  const handleAddToDeck = async (cardId: string, deckId: string) => {
    try {
      await deckClient.addCardToDeck({ cardId, deckId }, createOptions(token));
      setActiveCardIdDropdown(null);
      // Refresh decks to update card inclusion status
      fetchData();
    } catch (err: any) {
      alert(`Failed to add to deck: ${err.message || err}`);
    }
  };

  return (
    <div className="max-w-6xl mx-auto p-4 py-12">
      <header className="flex justify-between items-start mb-12">
        <div>
          <h1 className="text-5xl font-black tracking-tighter text-white">CARDS</h1>
          <p className="text-gray-400 font-bold uppercase tracking-widest text-sm">
            Manage your authored content
          </p>
          {user && (
            <div className="mt-4 flex items-center gap-2">
              <div className="w-8 h-8 rounded-full bg-primary/20 flex items-center justify-center text-primary text-xs font-bold ring-1 ring-primary/30">
                {user.name.slice(0, 2).toUpperCase()}
              </div>
              <span className="text-gray-300 font-bold text-sm tracking-tight">{user.name}</span>
            </div>
          )}
        </div>
        <div className="flex gap-3">
          <button
            onClick={handleOpenCardDialog}
            className="bg-secondary hover:bg-secondary/80 text-black font-black px-6 py-3 rounded-2xl flex items-center gap-2 transition-all transform hover:scale-105 shadow-lg shadow-secondary/10"
          >
            <Plus size={20} />
            NEW CARD
          </button>
          <button
            onClick={logout}
            className="p-3 bg-white/5 hover:bg-white/10 text-gray-400 hover:text-white rounded-2xl transition-all ring-1 ring-white/5"
            title="Logout"
          >
            <LogOut size={20} />
          </button>
        </div>
      </header>

      <div className="glass p-8 rounded-[2.5rem] min-h-[500px] border border-white/5 shadow-2xl relative overflow-hidden">
        {loading && (
           <div className="absolute inset-0 bg-background/50 backdrop-blur-sm z-10 flex items-center justify-center">
              <div className="animate-spin rounded-full h-12 w-12 border-t-2 border-primary"></div>
           </div>
        )}

        <div className="space-y-8">
          <div className="flex flex-col md:flex-row md:items-end justify-between gap-6">
            <div>
              <h2 className="text-3xl font-black text-white tracking-tight">Authored Content</h2>
              <p className="text-gray-500 text-sm font-bold uppercase tracking-widest">The results of your dark imagination</p>
            </div>

            <div className="flex flex-wrap items-center gap-4">
              {/* Search Filter */}
              <div className="relative group/search min-w-[240px]">
                <Search className="absolute left-4 top-1/2 -translate-y-1/2 text-gray-500 group-hover/search:text-primary transition-colors" size={18} />
                <input
                  type="text"
                  placeholder="Filter cards..."
                  value={filter}
                  onChange={(e) => setFilter(e.target.value)}
                  className="w-full bg-white/5 border border-white/10 rounded-2xl py-3 pl-12 pr-4 text-white font-medium placeholder:text-gray-600 focus:outline-none focus:ring-2 focus:ring-primary/20 focus:border-primary/30 transition-all"
                />
              </div>

              {/* Sort Field */}
              <div className="relative group/sort min-w-[180px]">
                <ChevronDown className="absolute right-4 top-1/2 -translate-y-1/2 text-gray-500 pointer-events-none" size={16} />
                <select
                  value={sortField}
                  onChange={(e) => setSortField(Number(e.target.value))}
                  className="w-full appearance-none bg-white/5 border border-white/10 rounded-2xl py-3 pl-5 pr-12 text-white font-bold text-sm tracking-tight focus:outline-none focus:ring-2 focus:ring-primary/20 transition-all cursor-pointer hover:bg-white/10"
                >
                  <option value={CardOrderField.CREATED_AT} className="bg-background">Date Created</option>
                  <option value={CardOrderField.TEXT} className="bg-background">Card Text</option>
                </select>
              </div>

              {/* Direction Toggle */}
              <button
                onClick={() => setDescending(!descending)}
                className={`p-3 rounded-2xl transition-all flex items-center gap-2 font-bold text-xs tracking-widest border ${
                  descending 
                    ? 'bg-primary/20 border-primary/30 text-primary' 
                    : 'bg-white/5 border-white/10 text-gray-400 hover:text-white'
                }`}
              >
                <ArrowUpDown size={18} />
                {descending ? 'DESC' : 'ASC'}
              </button>
            </div>
          </div>

          {cards.length > 0 ? (
            <div className="columns-1 md:columns-2 lg:columns-3 gap-6 space-y-6">
              {cards.map(card => {
                const availableDecks = decks.filter(d => 
                  (d.ownerId === user?.id || (d.contributors || []).includes(user?.id || '')) &&
                  !(d.cardIds || []).includes(card.id)
                );
                
                return (
                  <div key={card.id} className={`break-inside-avoid p-6 rounded-2xl border ${
                    card.color === 1 
                      ? 'bg-black text-white border-white/10 shadow-lg shadow-black/50' 
                      : 'bg-white text-black border-black/5 shadow-xl'
                  } group relative`}>
                    <p className="text-lg font-bold leading-tight pr-8">{card.text}</p>
                    
                    {/* Save to Deck Dropdown */}
                    {activeCardIdDropdown === card.id && (
                      <div className="absolute inset-0 z-20 bg-background/95 backdrop-blur-md rounded-2xl p-4 flex flex-col border border-primary/20 animate-in fade-in zoom-in duration-200">
                        <div className="flex justify-between items-center mb-4 pb-2 border-b border-white/10">
                          <span className="text-[10px] font-black uppercase tracking-widest text-primary">Add to Deck</span>
                          <button onClick={() => setActiveCardIdDropdown(null)} className="text-gray-500 hover:text-white transition-colors">
                            <X size={16} />
                          </button>
                        </div>
                        <div className="flex-1 overflow-y-auto space-y-1 custom-scrollbar">
                          {availableDecks.length === 0 ? (
                            <div className="h-full flex items-center justify-center text-center p-4">
                              <p className="text-xs text-gray-500 font-bold italic">Already in all your decks or no decks available.</p>
                            </div>
                          ) : (
                            availableDecks.map(deck => (
                              <button
                                key={deck.id}
                                onClick={() => handleAddToDeck(card.id, deck.id)}
                                className="w-full text-left p-3 rounded-xl hover:bg-primary/10 text-white transition-all flex items-center justify-between group/deck"
                              >
                                <span className="text-sm font-bold truncate pr-2">{deck.name}</span>
                                <Check size={14} className="text-primary opacity-0 group-hover/deck:opacity-100 transition-opacity" />
                              </button>
                            ))
                          )}
                        </div>
                      </div>
                    )}

                    <div className="mt-6 flex justify-between items-center opacity-40 group-hover:opacity-100 transition-all">
                      <div className="flex items-center gap-3">
                        <span className="text-[10px] font-black tracking-widest uppercase">
                          {card.color === 1 ? 'Black Card' : 'White Card'}
                        </span>
                        {user && (
                          <button 
                            onClick={() => setActiveCardIdDropdown(card.id)}
                            className="p-1.5 hover:bg-primary/10 hover:text-primary rounded-lg transition-all"
                            title="Save to Deck"
                          >
                            <FolderPlus size={16} />
                          </button>
                        )}
                      </div>
                      {user && card.ownerId === user.id && (
                        <button 
                          onClick={() => handleDelete(card.id)}
                          className="p-1.5 hover:bg-red-500/10 hover:text-red-500 transition-all"
                          title="Delete Card"
                        >
                          <Trash2 size={16} />
                        </button>
                      )}
                    </div>
                  </div>
                );
              })}
            </div>
          ) : (
            <div className="flex flex-col items-center justify-center py-20 text-center">
              <div className="p-4 bg-white/5 rounded-full mb-4">
                <Layers size={48} className="text-gray-700" />
              </div>
              <p className="text-gray-500 font-bold italic mb-6">No cards created yet. Write something terrible!</p>
            </div>
          )}

          {totalPages > 1 && (
            <div className="mt-12 flex justify-center items-center gap-6">
              <button
                onClick={() => setPageNumber(p => Math.max(1, p - 1))}
                disabled={pageNumber === 1 || loading}
                className="px-6 py-2 rounded-xl bg-white/5 border border-white/10 text-white font-bold disabled:opacity-30 disabled:cursor-not-allowed hover:bg-white/10 transition-colors"
              >
                PREVIOUS
              </button>
              <div className="flex flex-col items-center">
                <span className="text-white font-black text-xl tracking-tighter">{pageNumber}</span>
                <span className="text-gray-500 text-[10px] font-bold uppercase tracking-widest">OF {totalPages}</span>
              </div>
              <button
                onClick={() => setPageNumber(p => Math.min(totalPages, p + 1))}
                disabled={pageNumber === totalPages || loading}
                className="px-6 py-2 rounded-xl bg-white/5 border border-white/10 text-white font-bold disabled:opacity-30 disabled:cursor-not-allowed hover:bg-white/10 transition-colors"
              >
                NEXT
              </button>
            </div>
          )}
        </div>
      </div>
      <CreationDialog
        isOpen={isDialogOpen}
        onClose={() => setIsDialogOpen(false)}
        onCreate={handleCreate}
        title="Scribe New Card"
        placeholder="Enter the text of your misery (use ___ for blanks)..."
        label="Card Text"
        submitLabel="Create Card"
        maxLength={limits?.maxCardTextLength}
      />
    </div>
  );
};
