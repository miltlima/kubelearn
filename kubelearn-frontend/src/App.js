import React, { useState, useEffect } from 'react';


function App() {
  const [questions, setQuestions] = useState([]);
  const [quizStarted, setQuizStarted] = useState(false);
  const [quizFinished, setQuizFinished] = useState(false);
  const [score, setScore] = useState(0);
  const [elapsedTime, setElapsedTime] = useState(0);

  const startQuiz = () => {
    setQuizStarted(true);
    setQuizFinished(false);
    setElapsedTime(0);
    fetchQuestions();
  };

  const fetchQuestions = async () => {
    try {
      const response = await fetch('http://localhost:8083/questions');
      if (!response.ok) {
        throw new Error('Network response was not ok');
      }
      const data = await response.json();
      setQuestions(data);
    } catch (error) {
      console.error('Error fetching questions:', error);
    }
  };

  const finishQuiz = async () => {
    try {
      const response = await fetch('http://localhost:8083/finish');
      const data = await response.json();
      setScore(Math.round(data.score));
      setQuizFinished(true);
    } catch (error) {
      console.error('Error finishing quiz:', error);
    }
  };

  const resetQuiz = () => {
    setQuizStarted(false);
    setQuizFinished(false);
    setQuestions([]);
    setScore(0);
    setElapsedTime(0);
  };

  useEffect(() => {
    let timer;
    if (quizStarted && !quizFinished) {
      timer = setInterval(() => {
        setElapsedTime(prevTime => prevTime + 1);
      }, 1000);
    }
    return () => clearInterval(timer);
  }, [quizStarted, quizFinished]);

  const getDifficultyColor = (difficulty) => {
    let colorClass = '';
    if (difficulty === 'Easy') colorClass = 'text-green-500';
    if (difficulty === 'Medium') colorClass = 'text-yellow-500';
    if (difficulty === 'Hard') colorClass = 'text-red-500';
    return `font-bold ${colorClass}`;
  };

  return (
    <div className="min-h-screen bg-gray-100 flex flex-col items-center justify-center p-4 font-roboto-condensed">
      {/* Logo */}
      <img
        src="images/kubelearn_logo.png"
        alt="KubeLearn Logo"
        className="w-[400px] sm:w-[480px] md:w-[560px] lg:w-[640px] mb-10 mx-auto"
      />
        <h1 className="text-3xl font-bold text-gray-800 mb-20">“Kubernetes feels like magic… until it breaks. Let’s learn to tame the YAML together. Test your Kubernetes superpowers</h1>
      <div className="flex flex-col items-center w-full max-w-7xl">
        {!quizStarted && (
          <button
            onClick={startQuiz}
            className="bg-green-500 hover:bg-green-700 text-white font-bold py-2 px-4 rounded"
          >
            Start Quiz
          </button>
        )}

        {quizStarted && !quizFinished && (
          <div className="w-full">
            <div className="w-full flex flex-col">
              <div className="text-lg font-bold text-gray-800 mb-4">
                Time Elapsed: {Math.floor(elapsedTime / 60)}:
                {elapsedTime % 60 < 10 ? `0${elapsedTime % 60}` : elapsedTime % 60}
              </div>

              <div className="flex-1">
                <table className="min-w-full bg-white shadow-md rounded-lg overflow-hidden h-full">
                  <thead className="bg-gray-200 text-gray-600 uppercase text-sm leading-normal">
                    <tr>
                      <th className="py-3 px-6 text-left">Question</th>
                      <th className="py-3 px-6 text-left">Difficulty</th>
                    </tr>
                  </thead>
                  <tbody className="text-gray-600 text-sm font-light">
                    {questions.map((question, index) => (
                      <tr key={index} className="border-b border-gray-200 hover:bg-gray-100">
                        <td className="py-3 px-6 text-left font-bold">{question.TestName}</td>
                        <td className={`py-3 px-6 text-left ${getDifficultyColor(question.Difficulty)}`}>
                          {question.Difficulty}
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
              <button
                onClick={finishQuiz}
                className="bg-red-500 hover:bg-red-700 text-white font-bold py-2 px-4 rounded mt-4"
              >
                Check Results
              </button>
            </div>
          </div>
        )}

        {quizFinished && (
          <div className="text-center mt-6 w-full">
            <h2 className="text-2xl font-semibold text-gray-800">Results</h2>
            <p className="text-gray-700">Score: {score}%</p>
            {score < 85 ? (
              <p className="text-red-600 font-bold mt-4">
                **You did not reach the minimum score to pass. Keep studying!**
              </p>
            ) : (
              <p className="text-green-600 font-bold mt-4">
                **Congratulations! You passed! Keep studying!**
              </p>
            )}
            <table className="min-w-full bg-white shadow-md rounded-lg overflow-hidden mt-4">
              <thead className="bg-gray-200 text-gray-600 uppercase text-sm leading-normal">
                <tr>
                  <th className="py-3 px-6 text-left">Question</th>
                  <th className="py-3 px-6 text-left">Difficulty</th>
                  <th className="py-3 px-6 text-left">Result</th>
                </tr>
              </thead>
              <tbody className="text-gray-600 text-sm font-light">
                {questions.map((question, index) => (
                  <tr key={index} className="border-b border-gray-200 hover:bg-gray-100">
                    <td className="py-3 px-6 text-left font-bold">{question.TestName}</td>
                    <td className={`py-3 px-6 text-left ${getDifficultyColor(question.Difficulty)}`}>
                      {question.Difficulty}
                    </td>
                    <td className="py-3 px-6 text-left">
                      {question.Passed ? '✅' : '❌'}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>

            <div className="mt-6">
              <button
                onClick={startQuiz}
                className="bg-green-500 hover:bg-green-700 text-white font-bold py-2 px-4 rounded mr-4"
              >
                Start Again
              </button>
              <button
                onClick={resetQuiz}
                className="bg-red-500 hover:bg-red-700 text-white font-bold py-2 px-4 rounded"
              >
                Finish
              </button>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}

export default App;