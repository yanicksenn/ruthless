import React, { useRef, useEffect, useState, useCallback } from 'react';
import { X, UserPlus, Loader2, Users, Search } from 'lucide-react';
import { friendClient, createOptions } from '../api/client';
import { Player } from '../api/ruthless';
import { useAuth } from '../context/AuthContext';

interface InviteFriendDialogProps {
  isOpen: boolean;
  onClose: () => void;
  onAction: (identifier: string) => void;
  excludeFromSessionId?: string;
  excludeFromDeckId?: string;
  title?: string;
  buttonText?: string;
}

export const InviteFriendDialog: React.FC<InviteFriendDialogProps> = ({
  isOpen,
  onClose,
  onAction,
  excludeFromSessionId,
  excludeFromDeckId,
  title,
  buttonText
}) => {
  const dialogRef = useRef<HTMLDialogElement>(null);
  const { token } = useAuth();
  const [friends, setFriends] = useState<Player[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [page, setPage] = useState(1);
  const [totalCount, setTotalCount] = useState(0);
  const [searchTerm, setSearchTerm] = useState('');
  const [debouncedSearch, setDebouncedSearch] = useState('');

  // Debounce search term
  useEffect(() => {
    const timer = setTimeout(() => {
      setDebouncedSearch(searchTerm);
    }, 300);
    return () => clearTimeout(timer);
  }, [searchTerm]);

  useEffect(() => {
    const dialog = dialogRef.current;
    if (!dialog) return;
    if (isOpen) {
      dialog.showModal();
      setPage(1);
      setSearchTerm('');
      setDebouncedSearch('');
      fetchFriends(1, '');
    } else {
      dialog.close();
      setPage(1);
      setSearchTerm('');
      setDebouncedSearch('');
    }
  }, [isOpen]);

  useEffect(() => {
    if (isOpen) {
      fetchFriends(page, debouncedSearch);
    }
  }, [page, debouncedSearch, isOpen]);

  const fetchFriends = async (pageNum: number, currentFilter: string) => {
    setLoading(true);
    setError('');
    try {
      const response = await friendClient.listFriends({
        pageSize: 5,
        pageNumber: pageNum,
        excludeFromSessionId,
        excludeFromDeckId,
        filter: currentFilter
      }, createOptions(token));
      setFriends(response.response.friends || []);
      setTotalCount(response.response.totalCount || 0);
    } catch (err: any) {
      setError('Failed to fetch friends.');
      console.error(err);
    } finally {
      setLoading(false);
    }
  };

  const handleClose = () => {
    onClose();
  };

  return (
    <dialog
      ref={dialogRef}
      onClose={handleClose}
      className="bg-transparent backdrop:bg-background/80 backdrop:backdrop-blur-sm p-0 m-auto"
    >
      <div className="glass p-8 rounded-[2.5rem] w-[90vw] md:w-[60vw] lg:w-[40vw] min-w-[320px] border border-white/10 shadow-2xl card-shadow flex flex-col max-h-[80vh]">
        <div className="flex justify-between items-center mb-6">
          <h2 className="text-2xl font-black text-white tracking-tight uppercase italic flex items-center gap-2">
            <Users size={24} className="text-primary" /> {title || 'Invite a Friend'}
          </h2>
          <button 
            onClick={handleClose}
            className="p-2 hover:bg-white/10 rounded-xl transition-colors text-gray-400 hover:text-white"
          >
            <X size={24} />
          </button>
        </div>

        <div className="relative mb-6">
          <div className="absolute inset-y-0 left-0 pl-4 flex items-center pointer-events-none">
            <Search size={18} className="text-gray-500" />
          </div>
          <input
            type="text"
            placeholder="Search friends..."
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            className="w-full bg-white/5 border border-white/10 rounded-2xl pl-11 pr-5 py-3 text-white focus:outline-none focus:border-primary/50 transition-colors placeholder:text-gray-600 font-bold"
          />
        </div>

        <div className="flex-1 overflow-y-auto pr-2 space-y-2 min-h-[200px]">
          {loading ? (
             <div className="flex items-center justify-center py-12">
               <Loader2 className="animate-spin text-primary" size={32} />
             </div>
          ) : error ? (
             <div className="text-center text-red-400 py-8 font-bold">{error}</div>
          ) : friends.length === 0 ? (
             <div className="text-center text-gray-500 py-12 font-bold italic">
                No friends available to invite.
             </div>
          ) : (
             friends.map((friend) => {
               return (
                 <div key={friend.id} className="glass p-4 rounded-2xl border border-white/5 flex justify-between items-center transition-all hover:bg-white/5">
                   <div className="truncate pr-2">
                     <p className="font-bold text-white truncate">{friend.name}</p>
                     <p className="text-xs text-gray-500 italic truncate font-mono tracking-tighter">#{friend.identifier}</p>
                   </div>
                    <button
                      onClick={() => onAction(`${friend.name}#${friend.identifier}`)}
                      className="bg-primary hover:bg-primary-dark text-black font-black px-4 py-2 rounded-xl flex items-center gap-2 transition-all transform hover:scale-[1.02] shadow-sm shadow-primary/20 uppercase tracking-widest text-xs ml-2 flex-shrink-0"
                    >
                      <UserPlus size={14} /> {buttonText || 'Invite'}
                    </button>
                 </div>
               );
             })
          )}
        </div>

        {totalCount > 5 && (
          <div className="flex justify-between items-center mt-4 pt-4 border-t border-white/10">
            <button
              onClick={() => setPage(p => Math.max(1, p - 1))}
              disabled={page === 1}
              className="text-xs font-bold uppercase tracking-widest text-gray-400 hover:text-white disabled:opacity-30 disabled:cursor-not-allowed"
            >
              Previous
            </button>
            <span className="text-xs text-gray-500 font-bold">
              Page {page} of {Math.ceil(totalCount / 5)}
            </span>
            <button
              onClick={() => setPage(p => Math.min(Math.ceil(totalCount / 5), p + 1))}
              disabled={page * 5 >= totalCount}
              className="text-xs font-bold uppercase tracking-widest text-gray-400 hover:text-white disabled:opacity-30 disabled:cursor-not-allowed"
            >
              Next
            </button>
          </div>
        )}
      </div>
    </dialog>
  );
};
