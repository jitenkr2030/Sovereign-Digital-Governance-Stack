import React, { useState, useCallback, useRef, useEffect } from 'react';
import './Slider.css';

export interface SliderProps {
  /** Current value */
  value: number;
  /** Minimum value */
  min?: number;
  /** Maximum value */
  max?: number;
  /** Step size */
  step?: number;
  /** Disabled state */
  disabled?: boolean;
  /** Show value label */
  showValue?: boolean;
  /** Callback when value changes */
  onChange?: (value: number) => void;
  /** Custom class name */
  className?: string;
}

/**
 * Slider Component
 * 
 * Input component for selecting a value from a range.
 */
export const Slider: React.FC<SliderProps> = ({
  value,
  min = 0,
  max = 100,
  step = 1,
  disabled = false,
  showValue = true,
  onChange,
  className = '',
}) => {
  const trackRef = useRef<HTMLDivElement>(null);
  const [isDragging, setIsDragging] = useState(false);

  // Calculate percentage for positioning
  const percentage = ((value - min) / (max - min)) * 100;

  // Clamp value within range
  const clampValue = useCallback((val: number) => {
    return Math.min(Math.max(val, min), max);
  }, [min, max]);

  // Round to step
  const roundToStep = useCallback((val: number) => {
    return Math.round(val / step) * step;
  }, [step]);

  // Calculate value from position
  const getValueFromPosition = useCallback((position: number) => {
    if (!trackRef.current) return value;
    const rect = trackRef.current.getBoundingClientRect();
    const x = position - rect.left;
    const percent = Math.max(0, Math.min(1, x / rect.width));
    return roundToStep(min + percent * (max - min));
  }, [value, min, max, roundToStep]);

  // Handle mouse/touch move
  const handleMove = useCallback((clientX: number) => {
    if (disabled || !trackRef.current) return;
    const newValue = getValueFromPosition(clientX);
    onChange?.(clampValue(newValue));
  }, [disabled, onChange, getValueFromPosition, clampValue]);

  // Mouse event handlers
  const handleMouseDown = (e: React.MouseEvent) => {
    if (disabled) return;
    setIsDragging(true);
    handleMove(e.clientX);
  };

  useEffect(() => {
    if (isDragging) {
      const handleMouseMove = (e: MouseEvent) => handleMove(e.clientX);
      const handleMouseUp = () => setIsDragging(false);

      document.addEventListener('mousemove', handleMouseMove);
      document.addEventListener('mouseup', handleMouseUp);

      return () => {
        document.removeEventListener('mousemove', handleMouseMove);
        document.removeEventListener('mouseup', handleMouseUp);
      };
    }
  }, [isDragging, handleMove]);

  // Touch event handlers
  const handleTouchStart = (e: React.TouchEvent) => {
    if (disabled) return;
    setIsDragging(true);
    handleMove(e.touches[0].clientX);
  };

  useEffect(() => {
    if (isDragging) {
      const handleTouchMove = (e: TouchEvent) => handleMove(e.touches[0].clientX);
      const handleTouchEnd = () => setIsDragging(false);

      document.addEventListener('touchmove', handleTouchMove);
      document.addEventListener('touchend', handleTouchEnd);

      return () => {
        document.removeEventListener('touchmove', handleTouchMove);
        document.removeEventListener('touchend', handleTouchEnd);
      };
    }
  }, [isDragging, handleMove]);

  return (
    <div className={`slider-container ${disabled ? 'slider-disabled' : ''} ${className}`}>
      {showValue && (
        <span className="slider-value">{value.toLocaleString()}</span>
      )}
      <div
        ref={trackRef}
        className="slider-track"
        onMouseDown={handleMouseDown}
        onTouchStart={handleTouchStart}
      >
        <div
          className="slider-fill"
          style={{ width: `${percentage}%` }}
        />
        <div
          className={`slider-thumb ${isDragging ? 'slider-thumb-active' : ''}`}
          style={{ left: `${percentage}%` }}
          role="slider"
          aria-valuenow={value}
          aria-valuemin={min}
          aria-valuemax={max}
          tabIndex={disabled ? -1 : 0}
        />
      </div>
    </div>
  );
};

export default Slider;
