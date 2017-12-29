import { NgModule } from '@angular/core'
import { CommonModule } from '@angular/common'
import { BrowserModule } from '@angular/platform-browser'
import { FormsModule } from '@angular/forms'
import { HttpModule } from '@angular/http'

import { CollectionsModule } from './collections/collections.module'
import { ImportsModule } from './imports/imports.module'
import { ExportsModule } from './exports/exports.module'
import { ConfigsModule } from './configs/configs.module'
import { FilterModule } from './filter/filter.module'
import { BufferModule } from './buffer/buffer.module'
import { ToolsModule } from './tools/tools.module'

import { ApiService } from './api.service'

import { AppComponent } from './app.component'

@NgModule({
    imports: [
        CommonModule,
        HttpModule,
        BrowserModule,
        FormsModule,
        CollectionsModule,
        ImportsModule,
        ExportsModule,
        ConfigsModule,
        FilterModule,
        BufferModule,
        ToolsModule
    ],
    providers: [ApiService],
    declarations: [AppComponent],
    bootstrap: [AppComponent]
})

export class AppModule { }
