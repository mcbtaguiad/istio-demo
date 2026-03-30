import { render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import App from './App'

test('renders login page', () => {
	render(
		<MemoryRouter initialEntries={['/']}>
			<App />
		</MemoryRouter>
	)

	screen.debug()

	// expect(async screen.findByText(/backend version/i)).toBeInTheDocument()
	expect(screen.getByText(/login/i)).toBeInTheDocument()
})
