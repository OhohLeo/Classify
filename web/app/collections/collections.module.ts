import { NgModule } from '@angular/core'
import { CommonModule } from '@angular/common'
import { BrowserModule } from '@angular/platform-browser'
import { FormsModule } from '@angular/forms'
import { ItemModule } from '../item/item.module'

import { CollectionService } from './collection.service'
import { ApiService } from '../api.service'

import { CollectionsComponent } from './collections.component'
import { CreateCollectionComponent } from './create.component'
import { ModifyCollectionComponent } from './modify.component'
import { DisplayCollectionComponent } from './display.component'
import { DeleteCollectionComponent } from './delete.component'

@NgModule({
    imports: [CommonModule, BrowserModule, FormsModule, ItemModule],
    providers: [ApiService, CollectionService],
    declarations: [CollectionsComponent,
        CreateCollectionComponent,
        ModifyCollectionComponent,
        DisplayCollectionComponent,
        DeleteCollectionComponent],
    exports: [CollectionsComponent, DisplayCollectionComponent],
})

export class CollectionsModule { }
