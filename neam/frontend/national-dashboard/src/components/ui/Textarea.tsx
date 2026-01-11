import React, { forwardRef } from 'react';
import './Textarea.css';

export interface TextareaProps extends React.TextareaHTMLAttributes<HTMLTextAreaElement> {
  /** Label for the textarea */
  label?: string;
  /** Error message */
  error?: string;
  /** Helper text */
  helperText?: string;
  /** Show character count */
  showCount?: boolean;
  /** Maximum character count */
  maxLength?: number;
  /** Custom class name */
  className?: string;
}

/**
 * Textarea Component
 * 
 * Multi-line text input with label, error handling, and character count.
 */
export const Textarea = forwardRef<HTMLTextAreaElement, TextareaProps>(
  (
    {
      label,
      error,
      helperText,
      showCount = false,
      maxLength,
      className = '',
      id,
      ...props
    },
    ref
  ) => {
    const textareaId = id || `textarea-${Math.random().toString(36).substr(2, 9)}`;
    const hasError = !!error;

    return (
      <div className={`textarea-container ${className}`}>
        {(label || showCount) && (
          <div className="textarea-header">
            {label && (
              <label htmlFor={textareaId} className="textarea-label">
                {label}
              </label>
            )}
            {showCount && maxLength && (
              <span className="textarea-count">
                {props.value?.toString().length || 0}/{maxLength}
              </span>
            )}
          </div>
        )}
        
        <textarea
          ref={ref}
          id={textareaId}
          className={`textarea ${hasError ? 'textarea-error' : ''}`}
          maxLength={maxLength}
          aria-invalid={hasError}
          aria-describedby={error ? `${textareaId}-error` : helperText ? `${textareaId}-helper` : undefined}
          {...props}
        />
        
        {error && (
          <p id={`${textareaId}-error`} className="textarea-error-message">
            {error}
          </p>
        )}
        
        {helperText && !error && (
          <p id={`${textareaId}-helper`} className="textarea-helper">
            {helperText}
          </p>
        )}
      </div>
    );
  }
);

Textarea.displayName = 'Textarea';

export default Textarea;
