import React, { useEffect, useRef, useCallback, useState } from 'react';
import { createPortal } from 'react-dom';
import { Button } from '../Button';
import type { ButtonProps } from '../Button';
import './ActionModal.css';

export interface ModalSize {
  width?: string;
  height?: string;
}

export interface ActionModalProps {
  /** Modal title */
  title: string;
  /** Modal content */
  children: React.ReactNode;
  /** Whether the modal is open */
  isOpen: boolean;
  /** Callback when modal should close */
  onClose: () => void;
  /** Callback when confirm action is triggered */
  onConfirm?: () => void | Promise<void>;
  /** Confirm button text */
  confirmText?: string;
  /** Cancel button text */
  cancelText?: string;
  /** Confirm button variant */
  confirmVariant?: ButtonProps['variant'];
  /** Modal size preset or custom dimensions */
  size?: 'sm' | 'md' | 'lg' | 'xl' | 'full' | ModalSize;
  /** Disable confirm button while loading */
  isLoading?: boolean;
  /** Show close button in header */
  showCloseButton?: boolean;
  /** Show confirm button */
  showConfirmButton?: boolean;
  /** Show cancel button */
  showCancelButton?: boolean;
  /** Disable backdrop click to close */
  disableBackdropClick?: boolean;
  /** Custom class name */
  className?: string;
  /** Modal ID for accessibility */
  modalId?: string;
  /** Initial focus element selector */
  initialFocus?: string;
  /** Show danger style for critical actions */
  danger?: boolean;
  /** Additional footer content */
  footer?: React.ReactNode;
  /** Hide footer entirely */
  hideFooter?: boolean;
}

/**
 * ActionModal Component
 * 
 * A versatile modal dialog for confirmations and intervention actions.
 * Supports various sizes, loading states, and accessibility features.
 */
export const ActionModal: React.FC<ActionModalProps> = ({
  title,
  children,
  isOpen,
  onClose,
  onConfirm,
  confirmText = 'Confirm',
  cancelText = 'Cancel',
  confirmVariant = 'primary',
  size = 'md',
  isLoading = false,
  showCloseButton = true,
  showConfirmButton = true,
  showCancelButton = true,
  disableBackdropClick = false,
  className = '',
  modalId = 'action-modal',
  initialFocus = '[data-autofocus]',
  danger = false,
  footer,
  hideFooter = false,
}) => {
  const modalRef = useRef<HTMLDivElement>(null);
  const previousActiveElement = useRef<HTMLElement | null>(null);
  const [isConfirmLoading, setIsConfirmLoading] = useState(false);

  // Handle escape key
  const handleKeyDown = useCallback((event: KeyboardEvent) => {
    if (event.key === 'Escape') {
      onClose();
    }
  }, [onClose]);

  // Handle backdrop click
  const handleBackdropClick = useCallback((event: React.MouseEvent) => {
    if (disableBackdropClick) return;
    if (event.target === event.currentTarget) {
      onClose();
    }
  }, [onClose, disableBackdropClick]);

  // Handle confirm with loading state
  const handleConfirm = async () => {
    if (!onConfirm) return;
    
    setIsConfirmLoading(true);
    try {
      const result = onConfirm();
      if (result instanceof Promise) {
        await result;
      }
    } finally {
      setIsConfirmLoading(false);
    }
  };

  // Focus management
  useEffect(() => {
    if (isOpen) {
      previousActiveElement.current = document.activeElement as HTMLElement;
      
      // Focus the modal or initial focus element
      const modalElement = modalRef.current;
      if (modalElement) {
        const focusElement = modalElement.querySelector(initialFocus) as HTMLElement;
        if (focusElement) {
          focusElement.focus();
        } else {
          modalElement.focus();
        }
      }
      
      // Lock body scroll
      document.body.style.overflow = 'hidden';
      document.addEventListener('keydown', handleKeyDown);
    } else {
      // Restore body scroll
      document.body.style.overflow = '';
      document.removeEventListener('keydown', handleKeyDown);
      
      // Restore focus
      previousActiveElement.current?.focus();
    }

    return () => {
      document.body.style.overflow = '';
      document.removeEventListener('keydown', handleKeyDown);
    };
  }, [isOpen, handleKeyDown, initialFocus]);

  // Don't render if closed
  if (!isOpen) return null;

  // Determine size classes
  const getSizeClass = (): string => {
    if (typeof size === 'object') {
      return 'modal-custom';
    }
    return `modal-${size}`;
  };

  // Get custom styles if size is an object
  const getSizeStyles = (): React.CSSProperties => {
    if (typeof size === 'object') {
      return {
        width: size.width,
        height: size.height,
        maxWidth: '100vw',
        maxHeight: '100vh',
      };
    }
    return {};
  };

  // Portal to document body
  const modalContent = (
    <div
      className={`modal-overlay ${className}`}
      onClick={handleBackdropClick}
      role="presentation"
    >
      <div
        ref={modalRef}
        className={`modal-container ${getSizeClass()} ${danger ? 'modal-danger' : ''}`}
        style={getSizeStyles()}
        role="dialog"
        aria-modal="true"
        aria-labelledby={`${modalId}-title`}
        tabIndex={-1}
      >
        {/* Header */}
        <div className="modal-header">
          <h2 id={`${modalId}-title`} className="modal-title">
            {danger && (
              <span className="modal-danger-icon">
                <svg width="20" height="20" viewBox="0 0 24 24" fill="currentColor">
                  <path d="M12 2L1 21h22L12 2zm0 3.99L19.53 19H4.47L12 5.99zM11 10v4h2v-4h-2zm0 6v2h2v-2h-2z"/>
                </svg>
              </span>
            )}
            {title}
          </h2>
          {showCloseButton && (
            <button
              type="button"
              className="modal-close-button"
              onClick={onClose}
              aria-label="Close modal"
            >
              <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                <path d="M18 6L6 18M6 6l12 12" />
              </svg>
            </button>
          )}
        </div>

        {/* Body */}
        <div className="modal-body">
          {children}
        </div>

        {/* Footer */}
        {!hideFooter && (
          <div className="modal-footer">
            {footer}
            {showCancelButton && (
              <Button
                variant="ghost"
                onClick={onClose}
                disabled={isLoading || isConfirmLoading}
              >
                {cancelText}
              </Button>
            )}
            {showConfirmButton && (
              <Button
                variant={danger ? 'danger' : confirmVariant}
                onClick={handleConfirm}
                isLoading={isLoading || isConfirmLoading}
              >
                {confirmText}
              </Button>
            )}
          </div>
        )}
      </div>
    </div>
  );

  return createPortal(modalContent, document.body);
};

// Convenience component for simple confirmations
interface ConfirmModalProps extends Omit<ActionModalProps, 'children' | 'title'> {
  message: string;
}

export const ConfirmModal: React.FC<ConfirmModalProps> = ({
  message,
  ...props
}) => (
  <ActionModal {...props} title="Confirm Action">
    <p className="confirm-message">{message}</p>
  </ActionModal>
);

export default ActionModal;
