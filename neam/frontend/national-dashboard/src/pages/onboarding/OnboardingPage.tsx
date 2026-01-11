import React from 'react';
import { useNavigate } from 'react-router-dom';
import { Card } from '../../components/ui/Card';
import { Button } from '../../components/ui/Button';

/**
 * OnboardingPage Component
 *
 * First-time user onboarding flow.
 * Protected route requiring authentication.
 */
export const OnboardingPage: React.FC = () => {
  const navigate = useNavigate();
  const [currentStep, setCurrentStep] = React.useState(1);

  const steps = [
    {
      number: 1,
      title: 'Welcome',
      description: 'Get started with your new account',
    },
    {
      number: 2,
      title: 'Profile Setup',
      description: 'Tell us about yourself',
    },
    {
      number: 3,
      title: 'Preferences',
      description: 'Customize your experience',
    },
    {
      number: 4,
      title: 'Complete',
      description: 'You are ready to go!',
    },
  ];

  const handleNext = () => {
    if (currentStep < steps.length) {
      setCurrentStep(currentStep + 1);
    } else {
      navigate('/dashboard');
    }
  };

  const handleSkip = () => {
    navigate('/dashboard');
  };

  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50 py-12 px-4 sm:px-6 lg:px-8">
      <div className="max-w-2xl w-full space-y-8">
        {/* Progress Steps */}
        <div className="flex items-center justify-center space-x-4">
          {steps.map((step, index) => (
            <React.Fragment key={step.number}>
              <div className="flex flex-col items-center">
                <div
                  className={`w-10 h-10 rounded-full flex items-center justify-center font-bold ${
                    step.number <= currentStep
                      ? 'bg-primary-600 text-white'
                      : 'bg-gray-200 text-gray-500'
                  }`}
                >
                  {step.number < currentStep ? (
                    <svg
                      className="w-6 h-6"
                      fill="none"
                      viewBox="0 0 24 24"
                      stroke="currentColor"
                    >
                      <path
                        strokeLinecap="round"
                        strokeLinejoin="round"
                        strokeWidth={2}
                        d="M5 13l4 4L19 7"
                      />
                    </svg>
                  ) : (
                    step.number
                  )}
                </div>
                <span className="mt-2 text-xs text-gray-500 hidden sm:block">
                  {step.title}
                </span>
              </div>
              {index < steps.length - 1 && (
                <div
                  className={`w-12 h-1 ${
                    step.number < currentStep ? 'bg-primary-600' : 'bg-gray-200'
                  }`}
                />
              )}
            </React.Fragment>
          ))}
        </div>

        {/* Step Content */}
        <Card>
          <div className="p-8">
            {currentStep === 1 && (
              <div className="text-center">
                <h2 className="text-2xl font-bold text-gray-900">
                  Welcome to Our Platform!
                </h2>
                <p className="mt-4 text-gray-600">
                  We're excited to have you on board. This quick onboarding will
                  help you get set up in just a few minutes.
                </p>
                <div className="mt-8">
                  <svg
                    className="mx-auto h-32 w-32 text-primary-600"
                    fill="none"
                    viewBox="0 0 24 24"
                    stroke="currentColor"
                  >
                    <path
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      strokeWidth={1}
                      d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"
                    />
                  </svg>
                </div>
              </div>
            )}

            {currentStep === 2 && (
              <div>
                <h2 className="text-2xl font-bold text-gray-900">
                  Set Up Your Profile
                </h2>
                <p className="mt-2 text-gray-600">
                  Add some information to help others identify you.
                </p>
                <div className="mt-6 space-y-4">
                  <div>
                    <label className="block text-sm font-medium text-gray-700">
                      Display Name
                    </label>
                    <input
                      type="text"
                      className="mt-1 block w-full border-gray-300 rounded-md shadow-sm focus:ring-primary-500 focus:border-primary-500 sm:text-sm border px-3 py-2"
                      placeholder="How should we call you?"
                    />
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-gray-700">
                      Bio
                    </label>
                    <textarea
                      className="mt-1 block w-full border-gray-300 rounded-md shadow-sm focus:ring-primary-500 focus:border-primary-500 sm:text-sm border px-3 py-2"
                      rows={3}
                      placeholder="Tell us a bit about yourself..."
                    />
                  </div>
                </div>
              </div>
            )}

            {currentStep === 3 && (
              <div>
                <h2 className="text-2xl font-bold text-gray-900">
                  Customize Your Experience
                </h2>
                <p className="mt-2 text-gray-600">
                  Choose how you want to use the platform.
                </p>
                <div className="mt-6 space-y-4">
                  <div>
                    <label className="block text-sm font-medium text-gray-700">
                      Primary Use Case
                    </label>
                    <select className="mt-1 block w-full border-gray-300 rounded-md shadow-sm focus:ring-primary-500 focus:border-primary-500 sm:text-sm border px-3 py-2">
                      <option>Personal</option>
                      <option>Business</option>
                      <option>Education</option>
                    </select>
                  </div>
                  <div className="flex items-center">
                    <input
                      type="checkbox"
                      className="h-4 w-4 text-primary-600 focus:ring-primary-500 border-gray-300 rounded"
                    />
                    <label className="ml-2 block text-sm text-gray-700">
                      Enable email notifications
                    </label>
                  </div>
                  <div className="flex items-center">
                    <input
                      type="checkbox"
                      className="h-4 w-4 text-primary-600 focus:ring-primary-500 border-gray-300 rounded"
                    />
                    <label className="ml-2 block text-sm text-gray-700">
                      Join the community newsletter
                    </label>
                  </div>
                </div>
              </div>
            )}

            {currentStep === 4 && (
              <div className="text-center">
                <h2 className="text-2xl font-bold text-gray-900">
                  You're All Set!
                </h2>
                <p className="mt-4 text-gray-600">
                  Your account is ready. Explore the dashboard to get started.
                </p>
                <div className="mt-8">
                  <svg
                    className="mx-auto h-32 w-32 text-green-500"
                    fill="none"
                    viewBox="0 0 24 24"
                    stroke="currentColor"
                  >
                    <path
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      strokeWidth={1}
                      d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"
                    />
                  </svg>
                </div>
              </div>
            )}

            {/* Navigation Buttons */}
            <div className="mt-8 flex justify-between">
              <Button
                variant="ghost"
                onClick={handleSkip}
              >
                Skip for now
              </Button>
              <Button variant="primary" onClick={handleNext}>
                {currentStep < steps.length ? 'Continue' : 'Get Started'}
              </Button>
            </div>
          </div>
        </Card>
      </div>
    </div>
  );
};

export default OnboardingPage;
