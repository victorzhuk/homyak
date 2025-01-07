import { GithubLink } from './GithubLink.tsx'
import { Face } from './Face.tsx'
import { Feedback } from './Feedback.tsx'

import 'bootstrap/dist/css/bootstrap.min.css'

function App() {
  return (
    <>
      <div className="container">
        <div className="base row flex-row">
          <div className="flex-column col-12 col-md-6">
            <Face />
          </div>
          <div className="flex-column col-12 col-md-6">
            <Feedback />
          </div>
        </div>
      </div>
      <GithubLink />
    </>
  )
}

export default App
