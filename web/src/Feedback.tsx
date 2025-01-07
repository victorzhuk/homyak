export function Feedback() {
  return (
    <div className="d-flex justify-content-center">
      <a href="/feedback" target="_blank" className="btn contact btn-outline-secondary btn-lg">
        <svg id="i-mail" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 32 32" width="32" height="32" fill="none" stroke="currentcolor"
             strokeLinecap="round" strokeLinejoin="round" strokeWidth="2">
          <path d="M2 26 L30 26 30 6 2 6 Z M2 6 L16 16 30 6" />
        </svg>
        {' '}feedback me
      </a>
    </div>
  )
}