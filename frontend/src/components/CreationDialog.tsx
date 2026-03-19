import React, { useRef, useEffect, useState } from 'react';
import { X, PlusCircle, Check } from 'lucide-react';

interface CreationDialogProps {
  isOpen: boolean;
  onClose: () => void;
  onCreate: (value: string) => void;
  title: string;
  placeholder: string;
  label: string;
  submitLabel: string;
  maxLength?: number;
  minLength?: number;
  initialValue?: string;
}

export const CreationDialog: React.FC<CreationDialogProps> = ({
  isOpen,
  onClose,
  onCreate,
  title,
  placeholder,
  label,
  submitLabel,
  maxLength,
  minLength = 1,
  initialValue = ''
}) => {
  const dialogRef = useRef<HTMLDialogElement>(null);
  const [value, setValue] = useState(initialValue);

  // Update internal value when initialValue changes (when opening for edit)
  useEffect(() => {
    setValue(initialValue);
  }, [initialValue, isOpen]);

  const isInvalid = (maxLength && value.length > maxLength) || (value.trim().length < minLength && value.length > 0);

  useEffect(() => {
    const dialog = dialogRef.current;
    if (!dialog) return;

    if (isOpen) {
      dialog.showModal();
    } else {
      dialog.close();
    }
  }, [isOpen]);

  const handleClose = () => {
    setValue('');
    onClose();
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (value.trim()) {
      onCreate(value.trim());
      handleClose();
    }
  };

  const handleCancel = () => {
    handleClose();
  };

  return (
    <dialog
      ref={dialogRef}
      onClose={handleClose}
      className="bg-transparent backdrop:bg-background/80 backdrop:backdrop-blur-sm p-0 overflow-visible m-auto"
    >
      <div className="glass p-10 rounded-[2.5rem] w-[90vw] md:w-[60vw] lg:w-[40vw] min-w-[320px] border border-white/10 shadow-2xl card-shadow">
        <div className="flex justify-between items-center mb-8">
          <h2 className="text-3xl font-black text-white tracking-tight uppercase italic">{title}</h2>
          <button 
            onClick={handleCancel}
            className="p-2 hover:bg-white/10 rounded-xl transition-colors text-gray-400 hover:text-white"
          >
            <X size={24} />
          </button>
        </div>

        <form onSubmit={handleSubmit} className="space-y-6">
          <div className="space-y-2">
            <div className="flex justify-between items-end pl-1">
              <label className="text-gray-500 text-xs font-black uppercase tracking-widest">
                {label}
              </label>
              {maxLength && (
                <span className={`text-[10px] font-black ${value.length > maxLength ? 'text-red-400' : 'text-gray-500'}`}>
                  {value.length} / {maxLength}
                </span>
              )}
            </div>
            <textarea
              autoFocus
              value={value}
              onChange={(e) => setValue(e.target.value)}
              placeholder={placeholder}
              rows={3}
              className={`w-full bg-white/5 border rounded-2xl px-5 py-4 text-white focus:outline-none focus:ring-2 transition-all placeholder:text-gray-700 font-bold resize-none ${
                isInvalid ? 'border-red-500/50 focus:ring-red-500/20' : 'border-white/10 focus:ring-primary/50'
              }`}
            />
            {maxLength && value.length > maxLength && (
              <p className="text-[10px] text-red-400 font-black uppercase tracking-wider pl-1 font-mono">
                Exceeds maximum length of {maxLength}
              </p>
            )}
          </div>

          <div className="flex gap-3 pt-2">
            <button
              type="button"
              onClick={handleCancel}
              className="flex-1 px-6 py-4 rounded-2xl font-black text-gray-400 hover:text-white hover:bg-white/5 transition-all uppercase tracking-widest text-sm"
            >
              Cancel
            </button>
            <button
              type="submit"
              disabled={!value.trim() || isInvalid}
              className="flex-1 bg-primary hover:bg-primary-dark disabled:opacity-50 disabled:cursor-not-allowed text-black font-black px-6 py-4 rounded-2xl flex items-center justify-center gap-2 transition-all transform hover:scale-[1.02] shadow-lg shadow-primary/20 uppercase tracking-widest text-sm"
            >
              {initialValue ? <Check size={18} /> : <PlusCircle size={18} />}
              {submitLabel}
            </button>
          </div>
        </form>
      </div>
    </dialog>
  );
};
