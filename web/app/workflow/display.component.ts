import { Component, NgZone, Input, OnInit } from '@angular/core'
import { Reference } from '../references/reference'
import { BaseElement } from '../base'

declare var jQuery: any

@Component({
    selector: 'workflow-display',
    template: `
<div>
  <p>{{typeRef}}</p>
  <workflow-element *ngFor="let element of elementList"
                    [element]="element"
                    [ref]="typeRef">
  </workflow-element>
</div>
`,
    styles: [
`div {
  outline-color: red;
  outline-width: 4px;
  outline-style: solid;
}`]
})

export class DisplayComponent  implements OnInit {

    @Input() reference: Reference
    @Input() elements: Map<string, BaseElement[]>

    public typeRef: string
    public elementList: Array<BaseElement> = []

    constructor(private zone: NgZone) {}
    
    ngOnInit() {
	this.zone.run(() => {
	    this.typeRef = this.reference.getTypeRef()
	    this.elementList = this.elements.get(this.reference.ref)
	    console.log("LIST OF ELEMENTS?", this.typeRef, this.elementList)
	})
    }
}
