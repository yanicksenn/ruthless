import React, { useState, useEffect } from 'react';
import { useAuth } from '../context/AuthContext';
import { Plus, Trash2, LogOut } from 'lucide-react';
import { cardClient, createOptions } from '../api/client';
import { Card } from '../api/ruthless';
import { CreationDialog } from './CreationDialog';

export const Cards: React.FC = () => {
  const { token, user, logout } = useAuth();
  const [cards, setCards] = useState<Card[]>([]);
  const [loading, setLoading] = useState(true);
  const [isDialogOpen, setIsDialogOpen] = useState(false);
  const [pageNumber, setPageNumber] = useState(1);
  const [totalCount, setTotalCount] = useState(0);
  const pageSize = 12;

  const fetchData = async () => {
    setLoading(true);
    try {
      const cardsRes = await cardClient.listCards({ pageSize, pageNumber }, createOptions(token));
      setCards(cardsRes.response.cards || []);
      setTotalCount(cardsRes.response.totalCount);
    } catch (err) {
      console.error('Failed to fetch cards:', err);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchData();
  }, [token, pageNumber]);

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
          <div>
            <h2 className="text-3xl font-black text-white tracking-tight">Authored Content</h2>
            <p className="text-gray-500 text-sm font-bold uppercase tracking-widest">The results of your dark imagination</p>
          </div>

          <div className="columns-1 md:columns-2 lg:columns-3 gap-6 space-y-6">
            {cards.map(card => (
              <div key={card.id} className={`break-inside-avoid p-6 rounded-2xl border ${
                card.color === 1 
                  ? 'bg-black text-white border-white/10 shadow-lg shadow-black/50' 
                  : 'bg-white text-black border-black/5 shadow-xl'
              } group relative`}>
                <p className="text-lg font-bold leading-tight pr-8">{card.text}</p>
                <div className="mt-6 flex justify-between items-center opacity-40 group-hover:opacity-100 transition-opacity">
                  <span className="text-[10px] font-black tracking-widest uppercase">
                    {card.color === 1 ? 'Black Card' : 'White Card'}
                  </span>
                  {user && card.ownerId === user.id && (
                    <button 
                      onClick={() => handleDelete(card.id)}
                      className="text-gray-400 hover:text-red-500 transition-colors"
                    >
                      <Trash2 size={14} />
                    </button>
                  )}
                </div>
              </div>
            ))}
            {cards.length === 0 && (
              <div className="md:col-span-3 py-20 text-center">
                <p className="text-gray-500 font-bold italic">No cards found. Go ahead, write something terrible.</p>
              </div>
            )}
          </div>

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
      />
    </div>
  );
};
