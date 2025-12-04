import { getExpandableRowClasses, getExpandableCellClasses, getExpandableRowStyles, getExpandableRowRowClasses } from '../expandableRow'

describe('expandableRow helpers', () => {
    it('returns classes with padding when expanded', () => {
        const classes = getExpandableRowClasses(true, true)
        expect(classes).toContain('opacity-100')
        expect(classes).toContain('pl-12')
    })

    it('returns styles with Pod max height when expanded', () => {
        const styles = getExpandableRowStyles(true, 'Pod')
        expect(styles).toEqual({ maxHeight: 'calc(100vh - 250px)' })
    })

    it('returns cell classes with bottom border when expanded', () => {
        const classes = getExpandableCellClasses(true)
        expect(classes).toContain('border-b')
    })

    it('returns row classes toggling expanded state', () => {
        expect(getExpandableRowRowClasses(true)).toContain('bg-gray-800/30')
        expect(getExpandableRowRowClasses(false)).not.toContain('bg-gray-800/30')
    })
})
